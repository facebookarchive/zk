package zk

import (
	"github.com/go-zookeeper/jute/lib/go/jute"
)

type zkConn interface {
	GetData(path string) ([]byte, error)
	GetChildren(path string) ([]string, error)
	Close() error
	isAlive() bool
}

type pendingRequest struct {
	reply jute.RecordReader
	done  chan struct{}
}
