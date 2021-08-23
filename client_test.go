package zk

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/facebookincubator/zk/integration"
	"github.com/facebookincubator/zk/io"
	"github.com/facebookincubator/zk/testutils"
)

const defaultMaxRetries = 5

func TestClientRetryLogic(t *testing.T) {
	server, err := testutils.NewDefaultServer()
	if err != nil {
		t.Fatalf("error creating test server: %v", err)
	}
	defer server.Close()

	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    server.Addr().Network(),
		Ensemble:   server.Addr().String(),
	}

	children, err := client.GetChildren(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}

	if expected := []string{"test"}; !reflect.DeepEqual(expected, children) {
		t.Fatalf("getChildren error: expected %v, got %v", expected, children)
	}
}

func TestClientRetryLogicFails(t *testing.T) {
	server, err := testutils.NewDefaultServer()
	if err != nil {
		t.Fatalf("error creating test server: %v", err)
	}

	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    server.Addr().Network(),
		Ensemble:   server.Addr().String(),
	}

	// close server before client makes RPC call
	if err = server.Close(); err != nil {
		t.Fatalf("unexpected error closing server: %v", err)
	}

	_, err = client.GetChildren(context.Background(), "/")
	if err == nil || !errors.Is(err, errMaxRetries) {
		t.Fatalf("expected error: \"%v\", got error: \"%v\"", errMaxRetries, err)
	}
}

func TestClientContextCanceled(t *testing.T) {
	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    "tcp",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// expect the client not to retry when ctx is canceled
	if _, err := client.GetData(ctx, "/"); !errors.Is(err, ctx.Err()) {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
}

func TestErrorCodeHandling(t *testing.T) {
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
		MaxRetries: defaultMaxRetries,
		Network:    "tcp",
		Ensemble:   "127.0.0.1:2181",
	}

	// attempt to access node that does not exist
	_, err = client.GetChildren(context.Background(), "/nonexisting")

	// verify that the ZK server error has been processed properly and had no retries
	var ioError *io.Error
	if errors.Is(err, errMaxRetries) || !errors.As(err, &ioError) {
		t.Fatalf("unexpected error calling GetChildren: %v", err)
	}

}
