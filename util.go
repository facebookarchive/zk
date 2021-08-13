package zk

import (
	"bytes"
	"fmt"
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

	if err := enc.WriteBuffer(requestBytes); err != nil {
		return nil, err
	}
	if err := enc.WriteEnd(); err != nil {
		return nil, err
	}

	return sendBuf.Bytes(), nil
}
