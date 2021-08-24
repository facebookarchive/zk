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

// ReadRecord reads the request header and body depending on the opcode.
// It returns the serialized request header and body, or an error if it occurs.
func ReadRecord(dec *jute.BinaryDecoder) (*proto.RequestHeader, jute.RecordReader, error) {
	header := &proto.RequestHeader{}

	if err := dec.ReadRecord(header); err != nil {
		return nil, nil, fmt.Errorf("error reading RequestHeader: %w", err)
	}

	var req jute.RecordReader
	switch header.Type {
	case opGetData:
		req = &proto.GetDataRequest{}
	case opGetChildren:
		req = &proto.GetChildrenRequest{}
	default:
		return nil, nil, fmt.Errorf("unrecognized header type: %d", header.Type)
	}

	if err := dec.ReadRecord(req); err != nil {
		return nil, nil, fmt.Errorf("error reading request: %w", err)
	}

	return header, req, nil
}
