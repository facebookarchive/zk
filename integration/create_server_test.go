package integration

import (
	"testing"
	"time"

	"github.com/facebookincubator/zk/flw"
)

func TestCreateServer(t *testing.T) {
	cfg := DefaultConfig()

	server, err := NewZKServer("3.6.2", cfg)
	if err != nil {
		t.Errorf("unexpected error while initializing zk server: %v", err)
	}

	// run ZK in separate goroutine
	if err = server.Run(); err != nil {
		t.Errorf("unexpected error while calling RunZookeeperServer: %s", err)
		return
	}

	// verify server status is ok
	oks := flw.Ruok([]string{"0.0.0.0"}, 5*time.Second)
	if len(oks) < 1 || !oks[0] {
		t.Errorf("ruok indicates server is running in an error state")
	}
	server.Shutdown()
}
