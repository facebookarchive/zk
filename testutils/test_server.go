package testutils

import (
	"fmt"
	"net"

	"github.com/facebookincubator/zk/internal/proto"
	"github.com/facebookincubator/zk/io"
	"github.com/facebookincubator/zk/opcodes"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// defaultListenAddress is the default address on which the test server listens.
const defaultListenAddress = "127.0.0.1:"

// TestServer is a mock Zookeeper server which enables local testing without the need for a Zookeeper instance.
type TestServer struct {
	listener net.Listener
}

// NewServer creates a new TestServer instance with a default local listener.
func NewServer() (*TestServer, error) {
	l, err := newLocalListener()
	if err != nil {
		return nil, err
	}

	return &TestServer{listener: l}, nil
}

// Addr returns the addres on which this test server is listening on.
func (s *TestServer) Addr() net.Addr {
	return s.listener.Addr()
}

// Close closes the test server's listener.
func (s *TestServer) Close() error {
	return s.listener.Close()
}

// Start starts the test server and its handler.
// Calls to this method are blocking so calling in a separate goroutine is recommended.
func (s *TestServer) Start() {
	s.handler()
}

func newLocalListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", defaultListenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on a port: %v", err)
	}

	return listener, nil
}

func (s *TestServer) handler() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}

		go func() {
			if err := handleConn(conn); err != nil {
				return
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	if err := jute.NewBinaryDecoder(conn).ReadRecord(&proto.ConnectRequest{}); err != nil {
		return err
	}

	if err := serializeAndSend(conn, &proto.ConnectResponse{}); err != nil {
		return err
	}

	dec := jute.NewBinaryDecoder(conn)
	for {
		if _, err := dec.ReadInt(); err != nil {
			continue
		}

		header := &proto.RequestHeader{}
		if err := dec.ReadRecord(header); err != nil {
			return err
		}
		switch header.Type {
		case opcodes.OpGetData:
			resp := &proto.GetDataResponse{Data: []byte("test")}
			return serializeAndSend(conn, &proto.ReplyHeader{Xid: header.Xid}, resp)
		case opcodes.OpGetChildren:
			resp := &proto.GetChildrenResponse{Children: []string{"test"}}
			return serializeAndSend(conn, &proto.ReplyHeader{Xid: header.Xid}, resp)
		default:
			continue
		}
	}

}

func serializeAndSend(conn net.Conn, resp ...jute.RecordWriter) error {
	sendBuf, err := io.SerializeWriters(resp...)
	if err != nil {
		return err
	}
	if _, err = conn.Write(sendBuf); err != nil {
		return err
	}
	return nil
}
