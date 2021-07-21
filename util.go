package zk

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type pendingRequest struct {
	reply jute.RecordReader
	done  chan struct{}
}

// DialContext is a function to be used to establish a connection to a single host.
type DialContext func(ctx context.Context, network, addr string) (net.Conn, error)

type Client struct {
	Dialer  DialContext
	Timeout time.Duration
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
