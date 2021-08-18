package zk

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

const defaultMaxRetries = 5

var testError = fmt.Errorf("error")

// mockConnRPC is a mock implementation of the zkConn interface used for testing purposes.
type mockConnRPC struct {
	callCount              int
	retriesUntilFunctional int
}

func (c *mockConnRPC) isAlive() bool {
	return true
}

func (c *mockConnRPC) GetData(path string) ([]byte, error) {
	return []byte("mock"), nil
}

// GetChildren is a mock implementation which will return an error a given number of times to test the retry logic.
func (c *mockConnRPC) GetChildren(path string) ([]string, error) {
	if c.callCount == c.retriesUntilFunctional {
		return []string{"zookeeper"}, nil
	}
	c.callCount++

	return nil, testError
}

func (c *mockConnRPC) Close() error {
	return nil
}

func TestClientRetryLogic(t *testing.T) {
	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    "tcp",
		conn:       &mockConnRPC{retriesUntilFunctional: defaultMaxRetries},
	}
	defer client.Reset()

	expected := []string{"zookeeper"}
	children, err := client.GetChildren(context.Background(), "/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}

	if !reflect.DeepEqual(expected, children) {
		t.Fatalf("getChildren error: expected %v, got %v", expected, children)
	}
}

func TestClientRetryLogicFails(t *testing.T) {
	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    "tcp",
		conn:       &mockConnRPC{retriesUntilFunctional: defaultMaxRetries + 1},
	}
	defer client.Reset()

	expectedErr := fmt.Errorf("connection failed after %d retries: %w", defaultMaxRetries, testError)
	_, err := client.GetChildren(context.Background(), "/")
	if err == nil || err.Error() != expectedErr.Error() {
		t.Fatalf("expected error: \"%v\", got error: \"%v\"", err, expectedErr)
	}
}

func TestClientContextCanceled(t *testing.T) {
	client := &Client{
		MaxRetries: defaultMaxRetries,
		Network:    "tcp",
		conn:       &mockConnRPC{},
	}
	defer client.Reset()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// expect the client not to retry when ctx is canceled
	if _, err := client.GetData(ctx, "/"); !errors.Is(err, ctx.Err()) {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
}
