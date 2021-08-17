package testutils

import (
	"context"
	"net"
	"sync"

	"github.com/facebookincubator/zk/internal/proto"
	"github.com/facebookincubator/zk/io"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// NewListener returns a new TestListener instance with default values.
func NewListener() *TestListener {
	return &TestListener{
		connections: make(chan net.Conn),
		closed:      make(chan struct{}),
	}
}

// TestListener is an in-memory net.Listener implementation used for testing.
type TestListener struct {
	server      net.Conn
	connections chan net.Conn

	closeLock sync.Mutex
	closed    chan struct{}
}

// DialContext returns the client connection for the in-memory listener.
func (l *TestListener) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	select {
	case <-l.closed:
		return nil, net.ErrClosed
	default:
	}

	server, client := net.Pipe()
	l.connections <- server

	return client, nil
}

// Accept blocks until a client connects or an error occurs
func (l *TestListener) Accept() (net.Conn, error) {
	select {
	case newConn := <-l.connections:
		return newConn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

// Close closes the in-memory listener.
func (l *TestListener) Close() error {
	l.closeLock.Lock()
	defer l.closeLock.Unlock()

	close(l.closed)

	return nil
}

// Addr returns empty IP address for the in-memory listener.
func (l *TestListener) Addr() net.Addr {
	return &net.IPAddr{}
}

// Handler receives a request header and body from the client connection,
// and returns the response along with a reply header.
func (l *TestListener) Handler(req jute.RecordReader, resp jute.RecordWriter) error {
	if l.server == nil {
		return l.onInit(req, resp)
	}

	dec := jute.NewBinaryDecoder(l.server)
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

func (l *TestListener) onInit(req jute.RecordReader, resp jute.RecordWriter) error {
	server, err := l.Accept()
	if err != nil {
		return err
	}
	l.server = server

	if err = jute.NewBinaryDecoder(l.server).ReadRecord(req); err != nil {
		return err
	}

	return l.serializeAndSend(resp)
}

func (l *TestListener) serializeAndSend(resp ...jute.RecordWriter) error {
	sendBuf, err := io.SerializeWriters(resp...)
	if err != nil {
		return err
	}
	if _, err = l.server.Write(sendBuf); err != nil {
		return err
	}
	return nil
}
