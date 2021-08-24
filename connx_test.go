package zk_test

import (
	"bytes"
	"context"
	"testing"

	. "github.com/facebookincubator/zk"
	. "github.com/facebookincubator/zk/testutils"
)

func TestGetDataSimple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping local ZK mock tests for CI")
	}
	server, err := NewDefaultServer()
	if err != nil {
		t.Fatalf("error creating test server: %v", err)
	}
	defer server.Close()

	conn, err := DialContext(context.Background(), server.Addr().Network(), server.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error dialing server: %v", err)
	}
	defer conn.Close()

	res, err := conn.GetData("/")
	if err != nil {
		t.Fatalf("unexpected error calling GetData: %v", err)
	}
	if expected := []byte("test"); !bytes.Equal(expected, res) {
		t.Fatalf("expected %v got %v", expected, res)
	}
}
