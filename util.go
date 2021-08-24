package zk

import (
	"bytes"
	"fmt"

	"github.com/facebookincubator/zk/internal/proto"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

// SerializeWriters takes in one or more RecordWriter instances, and serializes them to a byte array
// while also prepending the total length of the structures to the beginning of the array.
func SerializeWriters(generated ...jute.RecordWriter) ([]byte, error) {
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

// GetRecord returns a jute.RecordReader (typically a request type)
// based on the opcode received from a request header.
func GetRecord(opcode int32) (jute.RecordReader, error) {
	switch opcode {
	case opGetData:
		return &proto.GetDataRequest{}, nil
	case opGetChildren:
		return &proto.GetChildrenRequest{}, nil
	default:
		return nil, fmt.Errorf("unrecognized opcode: %d", opcode)
	}
}
