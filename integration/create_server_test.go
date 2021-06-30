package integration

import (
	"testing"
	"time"

	"github.com/facebookincubator/zk/flw"
)

func TestCreateServer(t *testing.T) {
	exitChan := make(chan struct{}, 1)
	readyChan := make(chan struct{}, 1)

	// run ZK in separate goroutine
	go func() {
		if err := RunZookeeperServer("3.6.2", "default.cfg", readyChan, exitChan); err != nil {
			t.Errorf("unexpected error while calling RunZookeeperServer: %s", err)
			readyChan <- struct{}{} // stops test goroutine from waiting
			return
		}
	}()

	// wait for server to start w/ a set timeout deadline
	select {
	case <-readyChan:
		oks := flw.Ruok([]string{"0.0.0.0"}, 5*time.Second)
		if len(oks) < 1 || !oks[0] {
			t.Errorf("ruok indicates server is running in an error state")
		}
		exitChan <- struct{}{}
		break
	case <-time.After(time.Minute):
		t.Errorf("timeout")
	}
}
