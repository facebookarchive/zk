package zk

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/facebookincubator/zk/internal/proto"
	"github.com/facebookincubator/zk/io"
	"github.com/facebookincubator/zk/testutils"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

const defaultMaxRetries = 5

func TestClientRetryLogic(t *testing.T) {
	// Create a new handler which will make the test server return an error for a set number of tries.
	// We expect the client to recover from these errors and retry the RPC calls until success on the last try.
	failCalls := defaultMaxRetries
	server, err := testutils.Serve(func(conn net.Conn, dec jute.Decoder) error {
		if failCalls > 0 {
			failCalls--

			header := &proto.RequestHeader{}
			if err := dec.ReadRecord(header); err != nil {
				return fmt.Errorf("error reading RequestHeader: %w", err)
			}

			if err := dec.ReadRecord(&proto.GetChildrenRequest{}); err != nil {
				return fmt.Errorf("error reading request body: %w", err)
			}

			// return reply with error code
			sendBuf, err := io.SerializeWriters(&proto.ReplyHeader{Xid: header.Xid, Err: 1}, &proto.GetChildrenResponse{})
			if err != nil {
				return fmt.Errorf("reply serialization error: %w", err)
			}
			if _, err = conn.Write(sendBuf); err != nil {
				return fmt.Errorf("reply write error: %w", err)
			}

			return nil
		}

		// after set number of failures, return normal response
		return testutils.DefaultHandler(conn, dec)
	})

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
	server, err := testutils.ServeDefault()
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
