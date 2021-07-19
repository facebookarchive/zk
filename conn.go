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

	"github.com/facebookincubator/zk/flw"
	"github.com/facebookincubator/zk/internal/data"
	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

var ErrSessionExpired = errors.New("zk: session has been expired by the server")
var emptyPassword = make([]byte, 16)

type Connection struct {
	provider        HostProvider
	conn            net.Conn
	lastZxid        int64 // last seen ZxID, representing a Zookeeper transaction ID
	sessionTimeout  time.Duration
	pingInterval    time.Duration
	sessionID       int64  // represents a Zookeeper session assigned by the server to the client
	passwd          []byte // generated by the server along with the session ID
	server          string // the server IP to which the client is currently connected
	xid             int32  // client-side request ID
	pendingRequests sync.Map
	cancelFunc      context.CancelFunc
}

// Connect connects the ZK client to the specified pool of Zookeeper servers with a desired timeout.
// The session will be considered valid after losing connection to the server based on the provided timeout.
func Connect(servers []string, timeout time.Duration) (*Connection, error) {
	conn := &Connection{
		provider:       &DNSHostProvider{},
		sessionTimeout: timeout,
		passwd:         emptyPassword,
	}
	conn.pingInterval = conn.sessionTimeout / 2

	err := conn.provider.Init(flw.FormatServers(servers))
	if err != nil {
		return nil, err
	}

	err = conn.dial()
	if err != nil {
		return nil, err
	}

	err = conn.authenticate()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	conn.cancelFunc = cancel

	go conn.handleReads(ctx)
	go conn.handleSession(ctx)

	return conn, nil
}

// Close closes the client connection, clearing all pending requests.
func (c *Connection) Close() error {
	c.cancelFunc()
	c.clearPendingRequests()

	return c.conn.Close()
}

func (c *Connection) dial() error {
	c.server, _ = c.provider.Next()
	conn, err := net.Dial("tcp", c.server)
	if err != nil {
		return fmt.Errorf("error dialing ZK server: %v", err)
	}

	c.conn = conn

	return nil
}

func (c *Connection) authenticate() error {
	// create and encode request for zk server
	request := &proto.ConnectRequest{
		ProtocolVersion: defaultProtocolVersion,
		LastZxidSeen:    c.lastZxid,
		TimeOut:         int32(c.sessionTimeout.Milliseconds()),
		SessionId:       c.sessionID,
		Passwd:          c.passwd,
	}

	sendBuf, err := serializeWriters(request)
	if err != nil {
		return fmt.Errorf("error serializing request: %v", err)
	}

	// send request payload via net.conn
	c.conn.Write(sendBuf)

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

	if response.SessionId == 0 {
		c.sessionID = 0
		c.passwd = emptyPassword
		c.lastZxid = 0
		return ErrSessionExpired
	}

	c.sessionID = response.SessionId
	c.sessionTimeout = time.Duration(response.TimeOut) * time.Millisecond
	c.passwd = response.Passwd

	return nil
}

func (c *Connection) GetData(path string) ([]byte, error) {
	header := &proto.RequestHeader{
		Xid:  c.getXid(),
		Type: opGetData,
	}
	request := &proto.GetDataRequest{
		Path:  path,
		Watch: false,
	}
	sendBuf, err := serializeWriters(header, request)
	if err != nil {
		return nil, fmt.Errorf("error serializing request: %v", err)
	}
	r := &proto.GetDataResponse{
		Data: nil,
		Stat: &data.Stat{},
	}
	pending := pendingRequest{
		reply: r,
		done:  make(chan struct{}, 1),
	}

	c.pendingRequests.Store(header.Xid, pending)

	c.conn.Write(sendBuf)

	select {
	case <-pending.done:
		return r.Data, nil
	case <-time.After(c.sessionTimeout):
		return nil, fmt.Errorf("got a timeout waiting on response for xid %d", header.Xid)
	}
}

func (c *Connection) handleReads(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			dec := jute.NewBinaryDecoder(c.conn)
			_, err := dec.ReadInt() // read response length
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

			value, present := c.pendingRequests.LoadAndDelete(replyHeader.Xid)
			if present {
				pending := value.(pendingRequest)
				if err = dec.ReadRecord(pending.reply); err != nil {
					log.Printf("could not decode response struct: %v", err)
					break
				}

				pending.done <- struct{}{}
			}
		}
	}
}

func (c *Connection) handleSession(ctx context.Context) {
	pingTicker := time.NewTicker(c.pingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			header := &proto.RequestHeader{
				Xid:  pingXID,
				Type: opPing,
			}
			sendBuf, _ := serializeWriters(header)
			c.conn.Write(sendBuf)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Connection) clearPendingRequests() {
	c.pendingRequests.Range(func(key, value interface{}) bool {
		c.pendingRequests.Delete(key)
		return true
	})
}

func (c *Connection) getXid() int32 {
	if c.xid == math.MaxInt32 {
		c.xid = 1
	}
	c.xid++

	return c.xid
}
