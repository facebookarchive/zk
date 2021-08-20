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

// HandlerFunc is the function the server uses to return a response to the client based on the request received.
// Note that custom handlers need to send a ReplyHeader before a response as per the Zookeeper protocol.
type HandlerFunc func(net.Conn, *proto.RequestHeader) error

// TestServer is a mock Zookeeper server which enables local testing without the need for a Zookeeper instance.
type TestServer struct {
	listener        net.Listener
	ResponseHandler HandlerFunc
}

// ServeDefault creates and starts a new TestServer instance with a default local listener and handler.
// Started servers should be closed by calling Close.
func ServeDefault() (*TestServer, error) {
	return Serve(DefaultHandler)
}

// Serve creates and starts a new TestServer instance with a custom handler.
// Started servers should be closed by calling Close.
func Serve(handler HandlerFunc) (*TestServer, error) {
	l, err := newLocalListener()
	if err != nil {
		return nil, err
	}
	server := &TestServer{listener: l, ResponseHandler: handler}
	go server.accept()

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

func (s *TestServer) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			return
		}

		go func() {
			if err = s.handleConn(conn); err != nil {
				log.Printf("connection handler error: %v", err)
			}
		}()
	}
}

func (s *TestServer) handleConn(conn net.Conn) error {
	defer conn.Close()

	if err := jute.NewBinaryDecoder(conn).ReadRecord(&proto.ConnectRequest{}); err != nil {
		return fmt.Errorf("error reading ConnectRequest: %w", err)
	}

	if err := serializeAndSend(conn, &proto.ConnectResponse{}); err != nil {
		return fmt.Errorf("error sending ConnectResponse: %w", err)
	}

	dec := jute.NewBinaryDecoder(conn)
	for {
		if _, err := dec.ReadInt(); err != nil {
			return fmt.Errorf("error reading request length: %w", err)
		}
		header := &proto.RequestHeader{}
		if err := dec.ReadRecord(header); err != nil {
			return fmt.Errorf("error reading RequestHeader: %w", err)
		}
		switch header.Type {
		case io.OpGetData:
			if err := dec.ReadRecord(&proto.GetDataRequest{}); err != nil {
				return fmt.Errorf("error reading request: %w", err)
			}
		case io.OpGetChildren:
			if err := dec.ReadRecord(&proto.GetChildrenRequest{}); err != nil {
				return fmt.Errorf("error reading request: %w", err)
			}
		default:
			return fmt.Errorf("unrecognized header type: %d", header.Type)
		}

		if err := s.ResponseHandler(conn, header); err != nil {
			return fmt.Errorf("request handler error: %w", err)
		}
	}
}

// DefaultHandler sends a default response based on the header received.
func DefaultHandler(conn net.Conn, header *proto.RequestHeader) error {
	var resp jute.RecordWriter
	switch header.Type {
	case io.OpGetData:
		resp = &proto.GetDataResponse{Data: []byte("test")}
	case io.OpGetChildren:
		resp = &proto.GetChildrenResponse{Children: []string{"test"}}
	}

	return serializeAndSend(conn, &proto.ReplyHeader{Xid: header.Xid}, resp)
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
