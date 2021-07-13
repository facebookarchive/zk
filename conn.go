package zk

import (
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	"github.com/facebookincubator/zk/flw"
	"github.com/facebookincubator/zk/internal/data"
	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

var ErrSessionExpired = errors.New("zk: session has been expired by the server")
var emptyPassword = make([]byte, 16)

type Connection struct {
	provider       HostProvider
	conn           net.Conn
	lastZxid       int64
	sessionTimeout time.Duration
	sessionID      int64
	passwd         []byte
	server         string
	xid            int32
}

func Connect(servers []string) (*Connection, error) {
	conn := &Connection{
		provider:       &DNSHostProvider{},
		sessionTimeout: time.Second,
		passwd:         emptyPassword,
	}

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

	return conn, nil
}

func (c *Connection) getXid() int32 {
	if c.xid == math.MaxInt32 {
		c.xid = 1
	}
	c.xid++

	return c.xid
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

	c.conn.Write(sendBuf)

	dec := jute.NewBinaryDecoder(c.conn)
	_, err = dec.ReadInt() // read response length
	if err != nil {
		return nil, fmt.Errorf("could not decode response length: %v", err)
	}
	replyHeader := &proto.ReplyHeader{}
	if err = dec.ReadRecord(replyHeader); err != nil {
		return nil, fmt.Errorf("could not decode response struct: %v", err)
	}
	response := &proto.GetDataResponse{
		Data: nil,
		Stat: &data.Stat{},
	}
	if err = dec.ReadRecord(response); err != nil {
		return nil, fmt.Errorf("could not decode response struct: %v", err)
	}

	return response.Data, nil
}
