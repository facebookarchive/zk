package zk

import (
	"github.com/go-zookeeper/jute/lib/go/jute"
)

type pendingRequest struct {
	reply jute.RecordReader
	done  chan struct{}
}
