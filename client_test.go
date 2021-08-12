package zk

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/facebookincubator/zk/integration"
)

func TestClientGetChildren(t *testing.T) {
	server, err := integration.NewZKServer("3.6.2", integration.DefaultConfig())
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

	client := &Client{
		Network:           "tcp",
		EnsembleAddresses: []string{"127.0.0.1:2180", "127.0.0.1:2181"},
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
	server, err := integration.NewZKServer("3.6.2", integration.DefaultConfig())
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

	client := &Client{
		Network:           "tcp",
		EnsembleAddresses: []string{"127.0.0.1:2180", "127.0.0.1:2181"}, // add a nonexistent first address
	}

	if _, err = client.GetData(context.Background(), "/"); err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
}

func TestClientContextCanceled(t *testing.T) {
	sessionCtx, cancelSession := context.WithCancel(context.Background())
	c, _ := net.Pipe()

	conn := &Conn{conn: c, sessionCtx: sessionCtx, cancelSession: cancelSession}
	client := &Client{
		Network:           "tcp",
		EnsembleAddresses: []string{"127.0.0.1:2181"},
		conn:              conn,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// expect the client not to retry when ctx is canceled
	if _, err := client.GetData(ctx, "/"); !errors.Is(err, ctx.Err()) {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
}
