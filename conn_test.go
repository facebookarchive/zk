package zk

import (
	"testing"

	"github.com/facebookincubator/zk/integration"
)

func TestAuthentication(t *testing.T) {
	// create server
	cfg := integration.DefaultConfig()

	server, err := integration.NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Errorf("unexpected error while initializing zk server: %v", err)
	}

	// run ZK in separate goroutine
	if err = server.Run(); err != nil {
		t.Errorf("unexpected error while calling RunZookeeperServer: %s", err)
		return
	}

	// attempt to authenticate against server
	_, err = Connect([]string{"127.0.0.1"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	server.Shutdown()

}
