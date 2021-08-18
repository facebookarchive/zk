package testutils

import (
	"fmt"
	"log"
	"net"

	"github.com/facebookincubator/zk/internal/proto"
	"github.com/facebookincubator/zk/io"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// defaultListenAddress is the default address on which the test server listens.
const defaultListenAddress = "127.0.0.1:"

// TestServer is a mock Zookeeper server which enables local testing without the need for a Zookeeper instance.
type TestServer struct {
	listener net.Listener
}

// NewServer creates and starts a new TestServer instance with a default local listener.
// Started servers should be closed by calling Close.
func NewServer() (*TestServer, error) {
	l, err := newLocalListener()
	if err != nil {
		return nil, err
	}
	server := &TestServer{listener: l}
	go server.handler()

	return server, nil
}

// Addr returns the address on which this test server is listening on.
func (s *TestServer) Addr() net.Addr {
	return s.listener.Addr()
}

// Close closes the test server's listener.
func (s *TestServer) Close() error {
	return s.listener.Close()
}

func newLocalListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", defaultListenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on a port: %w", err)
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
			if err = handleConn(conn); err != nil {
				log.Printf("handler error: %v", err)
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	if err := jute.NewBinaryDecoder(conn).ReadRecord(&proto.ConnectRequest{}); err != nil {
		return fmt.Errorf("error reading ConnectRequest: %w", err)
	}

	if err := serializeAndSend(conn, &proto.ConnectResponse{}); err != nil {
		return fmt.Errorf("error sending ConnectResponse: %w", err)
	}

	dec := jute.NewBinaryDecoder(conn)
	for {
		if _, err := dec.ReadInt(); err != nil {
			continue
		}

		header := &proto.RequestHeader{}
		if err := dec.ReadRecord(header); err != nil {
			return fmt.Errorf("error reading RequestHeader: %w", err)
		}
		switch header.Type {
		case io.OpGetData:
			resp := &proto.GetDataResponse{Data: []byte("test")}
			if err := serializeAndSend(conn, &proto.ReplyHeader{Xid: header.Xid}, resp); err != nil {
				return err
			}
		case io.OpGetChildren:
			resp := &proto.GetChildrenResponse{Children: []string{"test"}}
			if err := serializeAndSend(conn, &proto.ReplyHeader{Xid: header.Xid}, resp); err != nil {
				return err
			}
		default:
			continue
		}
	}
}

func serializeAndSend(conn net.Conn, resp ...jute.RecordWriter) error {
	sendBuf, err := io.SerializeWriters(resp...)
	if err != nil {
		return fmt.Errorf("reply serialization error: %w", err)
	}
	if _, err = conn.Write(sendBuf); err != nil {
		return fmt.Errorf("reply write error: %w", err)
	}

	return nil
}
