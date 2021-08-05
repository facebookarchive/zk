package zk

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// Conn represents a client connection to a Zookeeper server and parameters needed to handle its lifetime.
type Conn struct {
	conn net.Conn

	// client-side request ID
	xid int32

	// the client sends a requested timeout, the server responds with the timeout that it can give the client
	sessionTimeout time.Duration
	// the client sends pings to the server in this interval to keep the connection alive
	pingInterval time.Duration

	reqs          sync.Map
	cancelSession context.CancelFunc
	sessionCtx    context.Context
}

// DialContext connects to the ZK server using the default client.
func DialContext(ctx context.Context, network, address string) (*Conn, error) {
	defaultClient := Client{}
	defaultDialer := &net.Dialer{}
	defaultClient.Dialer = defaultDialer.DialContext

	return defaultClient.DialContext(ctx, network, address)
}

// DialContext connects the ZK client to the specified Zookeeper server.
// The provided context is used to determine the dial lifetime.
func (client *Client) DialContext(ctx context.Context, network, address string) (*Conn, error) {
	conn, err := client.Dialer(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("error dialing ZK server: %v", err)
	}

	sessionCtx, cancel := context.WithCancel(context.Background())
	c := &Conn{
		conn:           conn,
		sessionTimeout: defaultTimeout,
		cancelSession:  cancel,
		sessionCtx:     sessionCtx,
	}

	if client.DialTimeout != 0 {
		c.sessionTimeout = client.DialTimeout
	}
	if err = c.authenticate(); err != nil {
		return nil, fmt.Errorf("error authenticating with ZK server: %v", err)
	}

	go c.handleReads()
	go c.keepAlive()

	return c, nil
}

// Close closes the client connection, clearing all pending requests.
func (c *Conn) Close() error {
	c.cancelSession()
	c.clearPendingRequests()

	return c.conn.Close()
}

func (c *Conn) authenticate() error {
	// create and encode request for zk server
	request := &proto.ConnectRequest{
		TimeOut: int32(c.sessionTimeout.Milliseconds()),
	}

	sendBuf, err := serializeWriters(request)
	if err != nil {
		return fmt.Errorf("error serializing request: %v", err)
	}

	// send request payload via net.conn
	if _, err = c.conn.Write(sendBuf); err != nil {
		return fmt.Errorf("error writing authentication request to net.conn: %v", err)
	}

	// receive bytes from same socket, reading the message length first
	dec := jute.NewBinaryDecoder(c.conn)

	_, err = dec.ReadInt() // read response length
	if err != nil {
		return fmt.Errorf("could not decode response length: %v", err)
	}
	response := proto.ConnectResponse{}
	if err = response.Read(dec); err != nil {
		return fmt.Errorf("could not decode response struct: %v", err)
	}

	c.sessionTimeout = time.Duration(response.TimeOut) * time.Millisecond

	// set the ping interval to half of the session timeout, according to Zookeeper documentation
	c.pingInterval = c.sessionTimeout / 2

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
		Xid:  c.getXid(),
		Type: opcode,
	}

	sendBuf, err := serializeWriters(header, w)
	if err != nil {
		return fmt.Errorf("error serializing request: %v", err)
	}

	pending := &pendingRequest{
		reply: r,
		done:  make(chan struct{}, 1),
	}

	c.reqs.Store(header.Xid, pending)

	if _, err = c.conn.Write(sendBuf); err != nil {
		return fmt.Errorf("error writing request to net.conn: %w", err)
	}

	select {
	case <-pending.done:
		return nil
	case <-c.sessionCtx.Done():
		return fmt.Errorf("session closed: %w", c.sessionCtx.Err())
	case <-time.After(c.sessionTimeout):
		return fmt.Errorf("got a timeout waiting on response for xid %d", header.Xid)
	}
}

func (c *Conn) handleReads() {
	defer c.Close()
	dec := jute.NewBinaryDecoder(c.conn)
	for {
		select {
		case <-c.sessionCtx.Done():
			return
		default:
			_, err := dec.ReadInt() // read response length
			if errors.Is(err, net.ErrClosed) {
				return // don't make further attempts to read from closed connection, close goroutine
			}
			if err != nil {
				log.Printf("could not decode response length: %v", err)
				break
			}

			replyHeader := &proto.ReplyHeader{}
			if err = dec.ReadRecord(replyHeader); err != nil {
				log.Printf("could not decode response struct: %v", err)
				break
			}
			if replyHeader.Xid == pingXID {
				continue // ignore ping responses
			}

			value, ok := c.reqs.LoadAndDelete(replyHeader.Xid)
			if ok {
				pending := value.(*pendingRequest)
				if err = dec.ReadRecord(pending.reply); err != nil {
					log.Printf("could not decode response struct: %v", err)
					return
				}

				pending.done <- struct{}{}
			}
		}
	}
}

func (c *Conn) keepAlive() {
	pingTicker := time.NewTicker(c.pingInterval)
	defer pingTicker.Stop()
	defer c.Close()
	for {
		select {
		case <-pingTicker.C:
			header := &proto.RequestHeader{
				Xid:  pingXID,
				Type: opPing,
			}
			sendBuf, err := serializeWriters(header)
			if err != nil {
				log.Printf("error serializing ping request: %v", err)
				continue
			}
			if _, err = c.conn.Write(sendBuf); err != nil {
				log.Printf("error writing ping request to net.conn: %v", err)
				continue
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

func (c *Conn) getXid() int32 {
	if c.xid == math.MaxInt32 {
		c.xid = 1
	}
	c.xid++

	return c.xid
}
