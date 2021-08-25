package testutils

import (
	"fmt"
	"log"
	"net"

	"github.com/facebookincubator/zk"
	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// defaultListenAddress is the default address on which the test server listens.
const defaultListenAddress = "127.0.0.1:"

// HandlerFunc is the function the server uses to return a response to the client based on the request received.
// If an error is present, an error code should be returned.
// Note that custom handlers need to send a ReplyHeader before a response as per the Zookeeper protocol.
type HandlerFunc func(reader jute.RecordReader) (jute.RecordWriter, zk.Code)

// TestServer is a mock Zookeeper server which enables local testing without the need for a Zookeeper instance.
type TestServer struct {
	listener        net.Listener
	ResponseHandler HandlerFunc
}

// NewDefaultServer creates and starts a new TestServer instance with a default local listener and handler.
// Started servers should be closed by calling Close.
func NewDefaultServer() (*TestServer, error) {
	return NewServer(DefaultHandler)
}

// NewServer creates and starts a new TestServer instance with a custom handler.
// Started servers should be closed by calling Close.
func NewServer(handler HandlerFunc) (*TestServer, error) {
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
		header, req, err := zk.ReadRecord(dec)
		if err != nil {
			return fmt.Errorf("error reading request: %w", err)
		}

		response, errCode := s.ResponseHandler(req)
		send := []jute.RecordWriter{&proto.ReplyHeader{Xid: header.Xid, Err: int32(errCode)}}
		if errCode == 0 && response != nil {
			send = append(send, response)
		}

		if err = serializeAndSend(conn, send...); err != nil {
			return fmt.Errorf("error sending response: %w", err)
		}
	}
}

// DefaultHandler returns a default response based on the request received, with no error code.
func DefaultHandler(request jute.RecordReader) (jute.RecordWriter, zk.Code) {
	var resp jute.RecordWriter
	switch request.(type) {
	case *proto.GetDataRequest:
		resp = &proto.GetDataResponse{Data: []byte("test")}
	case *proto.GetChildrenRequest:
		resp = &proto.GetChildrenResponse{Children: []string{"test"}}
	}

	return resp, 0
}

func serializeAndSend(conn net.Conn, resp ...jute.RecordWriter) error {
	sendBuf, err := zk.WriteRecords(resp...)
	if err != nil {
		return fmt.Errorf("reply serialization error: %w", err)
	}
	if _, err = conn.Write(sendBuf); err != nil {
		return fmt.Errorf("reply write error: %w", err)
	}

	return nil
}
