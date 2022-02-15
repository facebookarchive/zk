/*
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 *
 */

package zk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

const defaultTimeout = 2 * time.Second
const overflowBitMask = 1<<31 - 1

// Conn represents a client connection to a Zookeeper server and parameters needed to handle its lifetime.
type Conn struct {
	conn net.Conn

	// client-side request ID
	xid int32
	// the client sends a requested timeout, the server responds with the timeout that it can give the client
	sessionTimeout time.Duration

	reqs          sync.Map
	cancelSession context.CancelFunc
	sessionCtx    context.Context

	writeRecordsChan chan *writeRecordRequest
}

type pendingRequest struct {
	reply jute.RecordReader
	done  chan struct{}
	error error
}

type writeRecordRequest struct {
	header *proto.RequestHeader
	writer jute.RecordWriter
}

// isAlive() checks the TCP connection is alive by reading from the sessionCtx channel.
func (c *Conn) isAlive() bool {
	select {
	case <-c.sessionCtx.Done():
		return false
	default:
		return true
	}
}

// DialContext connects to the ZK server using the default client.
func DialContext(ctx context.Context, network, address string) (*Conn, error) {
	defaultClient := Client{}
	return defaultClient.DialContext(ctx, network, address)
}

// DialContext connects the ZK client to the specified Zookeeper server.
// The provided context is used to determine the dial lifetime.
func (client *Client) DialContext(ctx context.Context, network, address string) (*Conn, error) {
	if client.Dialer == nil {
		defaultDialer := &net.Dialer{}
		client.Dialer = defaultDialer.DialContext
	}

	conn, err := client.Dialer(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("could not dial ZK server: %w", err)
	}

	sessionCtx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		conn:             conn,
		sessionTimeout:   defaultTimeout,
		cancelSession:    cancel,
		sessionCtx:       sessionCtx,
		writeRecordsChan: make(chan *writeRecordRequest, 10),
	}

	if client.SessionTimeout != 0 {
		c.sessionTimeout = client.SessionTimeout
	}
	if err = c.authenticate(); err != nil {
		return nil, fmt.Errorf("could not authenticate with ZK server: %w", err)
	}

	go c.handleReadWrites()
	go c.keepAlive()

	return c, nil
}

// Close closes the client connection, clearing all pending requests.
func (c *Conn) Close() error {
	c.cancelSession()
	c.clearPendingRequests()
	// close(c.writeRecordsChan)

	return c.conn.Close()
}

func (c *Conn) authenticate() error {
	// create and encode request for zk server
	request := &proto.ConnectRequest{
		TimeOut: int32(c.sessionTimeout.Milliseconds()),
	}

	if err := WriteRecords(c.conn, request); err != nil {
		return fmt.Errorf("could not write authentication request: %w", err)
	}

	// receive bytes from same socket, reading the message length first
	dec, err := createDecoder(c.conn)
	if err != nil {
		return fmt.Errorf("could not read auth response: %w", err)
	}

	response := proto.ConnectResponse{}
	if err := response.Read(dec); err != nil {
		return fmt.Errorf("could not decode authentication response: %w", err)
	}

	if response.TimeOut > 0 {
		c.sessionTimeout = time.Duration(response.TimeOut) * time.Millisecond
	}

	return nil
}

// GetData calls Get on a Zookeeper server's node using the specified path and returns the server's response.
func (c *Conn) GetData(path string) ([]byte, error) {
	request := &proto.GetDataRequest{Path: path}
	response := &proto.GetDataResponse{}

	if err := c.rpc(opGetData, request, response); err != nil {
		return nil, fmt.Errorf("error sending GetData request: %w", err)
	}

	return response.Data, nil
}

// GetChildren returns all children of a node at the given path, if they exist.
func (c *Conn) GetChildren(path string) ([]string, error) {
	request := &proto.GetChildrenRequest{Path: path}
	response := &proto.GetChildrenResponse{}

	if err := c.rpc(opGetChildren, request, response); err != nil {
		return nil, fmt.Errorf("error sending GetChildren request: %w", err)
	}

	return response.Children, nil
}

func (c *Conn) rpc(opcode int32, w jute.RecordWriter, r jute.RecordReader) error {
	header := &proto.RequestHeader{
		Xid:  c.nextXid(),
		Type: opcode,
	}

	pending := &pendingRequest{
		reply: r,
		done:  make(chan struct{}, 1),
	}

	c.reqs.Store(header.Xid, pending)

	c.writeRecordsChan <- &writeRecordRequest{
		header: header,
		writer: w,
	}

	select {
	case <-pending.done:
		return pending.error
	case <-c.sessionCtx.Done():
		return fmt.Errorf("session closed: %w", c.sessionCtx.Err())
	case <-time.After(c.sessionTimeout):
		return fmt.Errorf("got a timeout waiting on response for xid %d", header.Xid)
	}
}

func (c *Conn) handleReadWrites() {
	defer c.Close()
	for {
		if c.sessionCtx.Err() != nil {
			return
		}
		wr := <-c.writeRecordsChan
		if err := WriteRecords(c.conn, wr.header, wr.writer); err != nil {
			log.Printf("could not write rpc request: %v", err)
			continue
		}

		dec, err := createDecoder(c.conn)
		if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
			return // don't make further attempts to read from closed connection, close goroutine
		}
		if err != nil {
			log.Printf("could not read response packet: %v", err)
			return
		}

		replyHeader := &proto.ReplyHeader{}
		if err = dec.ReadRecord(replyHeader); err != nil {
			log.Printf("could not decode reply header: %v", err)
			return
		}
		if replyHeader.Xid == pingXID {
			continue // ignore ping responses
		}

		value, ok := c.reqs.LoadAndDelete(replyHeader.Xid)
		if !ok {
			log.Printf("no matching request found for xid %d", replyHeader.Xid)
			continue
		}

		pending := value.(*pendingRequest)
		if replyHeader.Err != 0 {
			code := Error(replyHeader.Err)
			pending.error = &code
		} else if err = dec.ReadRecord(pending.reply); err != nil {
			log.Printf("could not decode reply record: %v", err)
			return
		}

		pending.done <- struct{}{}
	}
}

func (c *Conn) keepAlive() {
	// set the ping interval to half of the session timeout, according to Zookeeper documentation
	pingTicker := time.NewTicker(c.sessionTimeout / 2)
	defer pingTicker.Stop()

	defer c.Close()
	for {
		select {
		case <-pingTicker.C:
			header := &proto.RequestHeader{
				Xid:  pingXID,
				Type: opPing,
			}
			c.writeRecordsChan <- &writeRecordRequest{
				header: header,
			}
		case <-c.sessionCtx.Done():
			return
		}
	}
}

func (c *Conn) clearPendingRequests() {
	c.reqs.Range(func(key, value interface{}) bool {
		c.reqs.Delete(key)
		return true
	})
}

func (c *Conn) nextXid() int32 {
	return atomic.AddInt32(&c.xid, 1) & overflowBitMask
}
