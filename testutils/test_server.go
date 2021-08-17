package testutils

import (
	"fmt"
	"net"

	"github.com/facebookincubator/zk/internal/proto"
	"github.com/facebookincubator/zk/io"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// DefaultListenAddress is the default address on which the test server listens.
const DefaultListenAddress = "127.0.0.1:"

// TestServer is a mock Zookeeper server which enables local testing without the need for a Zookeeper instance.
type TestServer struct {
	listener net.Listener
	conn     net.Conn
}

// NewServer creates a new TestServer instance with a default local listener.
func NewServer() (*TestServer, error) {
	l, err := newLocalListener()
	if err != nil {
		return nil, err
	}

	return &TestServer{listener: l}, nil
}

// Handler receives a request header and body from the client connection,
// and returns the response along with a reply header.
func (l *TestServer) Handler(req jute.RecordReader, resp jute.RecordWriter) error {
	if l.conn == nil {
		return l.onInit(req, resp)
	}

	dec := jute.NewBinaryDecoder(l.conn)
	if _, err := dec.ReadInt(); err != nil {
		return err
	}
	header := &proto.RequestHeader{}
	if err := dec.ReadRecord(header); err != nil {
		return err
	}
	if err := dec.ReadRecord(req); err != nil {
		return err
	}

	return l.serializeAndSend(&proto.ReplyHeader{Xid: header.Xid}, resp)
}

func (l *TestServer) Addr() net.Addr {
	return l.listener.Addr()
}

// Close closes the test server's listener.
func (l *TestServer) Close() error {
	return l.listener.Close()
}

func (l *TestServer) onInit(req jute.RecordReader, resp jute.RecordWriter) error {
	conn, err := l.listener.Accept()
	if err != nil {
		return err
	}
	l.conn = conn

	if err = jute.NewBinaryDecoder(l.conn).ReadRecord(req); err != nil {
		return err
	}

	return l.serializeAndSend(resp)
}

func (l *TestServer) serializeAndSend(resp ...jute.RecordWriter) error {
	sendBuf, err := io.SerializeWriters(resp...)
	if err != nil {
		return err
	}
	if _, err = l.conn.Write(sendBuf); err != nil {
		return err
	}
	return nil
}

func newLocalListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", DefaultListenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on a port: %v", err)
	}

	return listener, nil
}
