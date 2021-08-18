package zk

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/facebookincubator/zk/testutils"
)

const defaultMaxRetries = 5

func TestClientRetryLogic(t *testing.T) {
	server, err := testutils.NewServer()
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
		t.Fatalf("unexpected error calling GetChildren: %v", err)
	}

	if expected := []string{"test"}; !reflect.DeepEqual(expected, children) {
		t.Fatalf("getChildren error: expected %v, got %v", expected, children)
	}
}

func TestClientRetryLogicFails(t *testing.T) {
	server, err := testutils.NewServer()
	if err != nil {
		t.Fatalf("error creating test server: %v", err)
	}
	// close server before client makes RPC call
	if err = server.Close(); err != nil {
		t.Fatalf("unexpected error closing server: %v", err)
	}

	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    server.Addr().Network(),
		Ensemble:   server.Addr().String(),
	}

	_, err = client.GetChildren(context.Background(), "/")
	expectedErr := fmt.Errorf("connection failed after %d retries: %w", defaultMaxRetries, errors.Unwrap(err))
	if err == nil || err.Error() != expectedErr.Error() {
		t.Fatalf("expected error: \"%v\", got error: \"%v\"", expectedErr, err)
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
