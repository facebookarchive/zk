package zk

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/facebookincubator/zk/internal/data"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

const defaultTimeout = 2 * time.Second

type Client struct {
	// Dialer is a function to be used to establish a connection to a single host.
	Dialer  func(ctx context.Context, network, addr string) (net.Conn, error)
	Timeout time.Duration
}

// WorldACL produces an ACL list containing a single ACL which uses the
// provided permissions, with the scheme "world", and ID "anyone", which
// is used by ZooKeeper to represent any user at all.
func WorldACL(perms int32) []data.ACL {
	return []data.ACL{{Perms: perms, Id: data.Id{Scheme: "world", Id: "anyone"}}}
}

type pendingRequest struct {
	reply jute.RecordReader
	done  chan struct{}
}

// serializeWriters takes in one or more RecordWriter instances, and serializes them to a byte array
// while also prepending the total length of the structures to the beginning of the array.
func serializeWriters(generated ...jute.RecordWriter) ([]byte, error) {
	sendBuf := &bytes.Buffer{}
	enc := jute.NewBinaryEncoder(sendBuf)

	for _, generatedStruct := range generated {
		if err := generatedStruct.Write(enc); err != nil {
			return nil, fmt.Errorf("could not encode struct: %v", err)
		}
	}
	// copy encoded request bytes
	requestBytes := append([]byte(nil), sendBuf.Bytes()...)

	// use encoder to prepend request length to the request bytes
	sendBuf.Reset()
	enc.WriteBuffer(requestBytes)
	enc.WriteEnd()

	return sendBuf.Bytes(), nil
}
