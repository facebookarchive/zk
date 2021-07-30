package zk

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/facebookincubator/zk/integration"
)

func TestAuthentication(t *testing.T) {
	// create server
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Fatalf("unexpected error while initializing zk server: %v", err)
	}
	defer server.Shutdown()

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

	if conn.sessionID == 0 {
		t.Errorf("expected non-zero session ID")
	}
}

func TestGetDataWorks(t *testing.T) {
	// create server
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Fatalf("unexpected error while initializing zk server: %v", err)
	}
	defer server.Shutdown()
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
	log.Printf("getData response: %+v", res)
}

func TestGetDataNoTimeout(t *testing.T) {
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Fatalf("unexpected error while initializing zk server: %v", err)
	}
	defer server.Shutdown()
	if err = server.Run(); err != nil {
		t.Fatalf("unexpected error while calling RunZookeeperServer: %s", err)
		return
	}
	sessionTimeout := 5 * time.Second
	client := Client{
		Timeout: sessionTimeout,
	}
	conn, err := client.DialContext(context.Background(), "tcp", "127.0.0.1:2181")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer conn.Close()

	_, err = conn.GetData("/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
	// close conn artificially, subsequent calls to getData should not wait for timeout
	conn.Close()
	// call GetData again
	_, err = conn.GetData("/")

	select {
	case <-time.After(sessionTimeout):
		t.Fatalf("client should not wait for timeout if connection is closed")
	default:
		if errors.Is(errors.Unwrap(err), net.ErrClosed) {
			log.Printf("got ErrClosed as expected")
			return
		}
		if err != nil {
			t.Fatalf("unexpected error calling GetData: %v", err)
		}
	}

}
