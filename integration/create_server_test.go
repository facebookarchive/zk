package integration

import (
	"testing"
	"time"
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

	// wait for zookeeper server to start w/ a set timeout deadline
	select {
	case <-readyChan:
		// zk server started successfully or got an error, terminate it
		exitChan <- struct{}{}
		break
	case <-time.After(time.Minute):
		t.Errorf("timeout")
	}
}
