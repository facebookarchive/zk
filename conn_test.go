package zk

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/facebookincubator/zk/integration"
	"github.com/facebookincubator/zk/internal/proto"
	"github.com/facebookincubator/zk/testutils"
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
	server := testutils.NewServer()
	defer server.Close()

	go func() {
		if err := server.Handler(&proto.ConnectRequest{}, &proto.ConnectResponse{}); err != nil {
			t.Errorf("unexpected handler error: %v", err)
			return
		}
	}()

	client := Client{}
	conn, err := client.DialContext(context.Background(), server.Addr().Network(), server.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error dialing server: %v", err)
	}
	defer conn.Close()

	expected := []byte("test")
	go func() {
		if err = server.Handler(&proto.GetDataRequest{}, &proto.GetDataResponse{Data: expected}); err != nil {
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
