package zk

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
)

// mockConnRPC is a mock implementation of the zkConn interface used for testing purposes.
type mockConnRPC struct{}

func (r *mockConnRPC) isAlive() bool {
	return true
}

func (*mockConnRPC) GetData(path string) ([]byte, error) {
	return []byte("mock"), nil
}

func (*mockConnRPC) GetChildren(path string) ([]string, error) {
	return []string{"zookeeper"}, nil
}

func (*mockConnRPC) Close() error {
	return nil
}

type mockNetConn struct {
	net.Conn
}

func (m *mockNetConn) Read([]byte) (int, error) {
	return 1, nil
}

func (m *mockNetConn) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockNetConn) Close() error {
	return nil
}

func TestClientGetChildren(t *testing.T) {
	client := &Client{
		Network: "tcp",
		conn:    &mockConnRPC{},
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return &mockNetConn{}, nil
		},
	}

	expected := []string{"zookeeper"}
	children, err := client.GetChildren(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}

	if !reflect.DeepEqual(expected, children) {
		t.Fatalf("getChildren error: expected %v, got %v", expected, children)
	}
}

func TestClientGetData(t *testing.T) {
	client := &Client{
		Network: "tcp",
		conn:    &mockConnRPC{},
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return &mockNetConn{}, nil
		},
	}

	if _, err := client.GetData(context.Background(), "/"); err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
}

func TestClientContextCanceled(t *testing.T) {
	client := &Client{
		Network:           "tcp",
		EnsembleAddresses: []string{"127.0.0.1:2181"},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// expect the client not to retry when ctx is canceled
	if _, err := client.GetData(ctx, "/"); !errors.Is(err, ctx.Err()) {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
}
