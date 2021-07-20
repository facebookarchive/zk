package zk

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type pendingRequest struct {
	reply jute.RecordReader
	done  chan struct{}
}

// Dialer is a function to be used to establish a connection to a single host.
type Dialer func(network, address string, timeout time.Duration) (net.Conn, error)

// ConnOption represents a connection option which can be passed to the Connect function.
type ConnOption func(c *Connection)

func DialerOption(dialer Dialer) ConnOption {
	return func(c *Connection) {
		c.dialer = dialer
	}
}

func HostProviderOption(provider HostProvider) ConnOption {
	return func(c *Connection) {
		c.provider = provider
	}
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
