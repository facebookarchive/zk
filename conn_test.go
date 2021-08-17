package zk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/facebookincubator/zk/integration"
	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

func TestAuthentication(t *testing.T) {
	// create server
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Fatalf("unexpected error while initializing zk server: %v", err)
	}
	defer func(server *integration.ZKServer) {
		if err = server.Shutdown(); err != nil {
			t.Fatalf("unexpected error while shutting down zk server: %v", err)
		}
	}(server)

	// run ZK in separate goroutine
	if err = server.Run(); err != nil {
		t.Fatalf("unexpected error while calling RunZookeeperServer: %s", err)
		return
	}

	// attempt to authenticate against server
	conn, err := DialContext(context.Background(), "tcp", "127.0.0.1:2181")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer conn.Close()
}

func TestGetDataWorks(t *testing.T) {
	// create server
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Fatalf("unexpected error while initializing zk server: %v", err)
	}
	defer func(server *integration.ZKServer) {
		if err = server.Shutdown(); err != nil {
			t.Fatalf("unexpected error while shutting down zk server: %v", err)
		}
	}(server)
	if err = server.Run(); err != nil {
		t.Fatalf("unexpected error while calling RunZookeeperServer: %s", err)
		return
	}
	conn, err := DialContext(context.Background(), "tcp", "127.0.0.1:2181")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer conn.Close()

	res, err := conn.GetData("/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}

	t.Logf("getData response: %+v", res)
}

func TestGetDataNoTimeout(t *testing.T) {
	sessionCtx, cancelSession := context.WithCancel(context.Background())
	client, _ := net.Pipe()

	conn := Conn{conn: client, sessionCtx: sessionCtx, cancelSession: cancelSession}
	// close conn before sending request
	conn.Close()
	_, err := conn.GetData("/")
	select {
	case <-time.After(defaultTimeout):
		t.Fatalf("client should not wait for timeout if connection is closed")
	default:
		if err != nil && !errors.Is(errors.Unwrap(err), io.ErrClosedPipe) {
			t.Fatalf("unexpected error calling GetData: %v", err)
		}
	}
}
func TestGetChildrenDefault(t *testing.T) {
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Fatalf("unexpected error while initializing zk server: %v", err)
	}
	defer func(server *integration.ZKServer) {
		if err = server.Shutdown(); err != nil {
			t.Fatalf("unexpected error while shutting down zk server: %v", err)
		}
	}(server)
	if err = server.Run(); err != nil {
		t.Fatalf("unexpected error while calling RunZookeeperServer: %s", err)
		return
	}

	conn, err := DialContext(context.Background(), "tcp", "127.0.0.1:2181")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer conn.Close()

	expected := []string{"zookeeper"}
	res, err := conn.GetChildren("/")
	if err != nil {
		t.Fatalf("unexpected error calling GetChildren: %v", err)
	}

	if !reflect.DeepEqual(expected, res) {
		t.Fatalf("getChildren error: expected %v, got %v", expected, res)
	}
}

func TestGetDataSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in-memory tests for CI")
	}

	listener := NewListener()
	defer listener.Close()

	go func() {
		if err := listener.handler(&proto.ConnectRequest{}, &proto.ConnectResponse{}); err != nil {
			t.Errorf("unexpected handler error: %v", err)
			return
		}
	}()

	client := Client{Dialer: listener.DialContext}
	conn, err := client.DialContext(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error dialing server: %v", err)
	}
	defer conn.Close()

	expected := []byte("test")
	go func() {
		if err = listener.handler(&proto.GetDataRequest{}, &proto.GetDataResponse{Data: expected}); err != nil {
			t.Errorf("unexpected handler error: %v", err)
			return
		}
	}()

	res, err := conn.GetData("/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
	if !bytes.Equal(expected, res) {
		t.Fatalf("expected %v got %v", expected, res)
	}
}

func NewListener() *inMemoryListener {
	return &inMemoryListener{
		connections: make(chan net.Conn),
		closed:      make(chan struct{}),
	}
}

type inMemoryListener struct {
	server      net.Conn
	connections chan net.Conn

	closeLock sync.Mutex
	closed    chan struct{}
}

func (l *inMemoryListener) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
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
func (l *inMemoryListener) Accept() (net.Conn, error) {
	select {
	case newConn := <-l.connections:
		return newConn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

// Close closes the in-memory listener.
func (l *inMemoryListener) Close() error {
	l.closeLock.Lock()
	defer l.closeLock.Unlock()

	close(l.closed)

	return nil
}

// Addr returns empty IP address for the in-memory listener.
func (l *inMemoryListener) Addr() net.Addr {
	return &net.IPAddr{}
}

func (l *inMemoryListener) handler(req jute.RecordReader, resp jute.RecordWriter) error {
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

func (l *inMemoryListener) onInit(req jute.RecordReader, resp jute.RecordWriter) error {
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

func (l *inMemoryListener) serializeAndSend(resp ...jute.RecordWriter) error {
	sendBuf, err := serializeWriters(resp...)
	if err != nil {
		return err
	}
	if _, err = l.server.Write(sendBuf); err != nil {
		return err
	}
	return nil
}
