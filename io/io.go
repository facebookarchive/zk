package io

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

var errNilStruct = errors.New("cannot generate nil struct")

// SerializeWriters takes in one or more RecordWriter instances, and serializes them to a byte array
// while also prepending the total length of the structures to the beginning of the array.
func SerializeWriters(generated ...jute.RecordWriter) ([]byte, error) {
	if generated == nil {
		return nil, errNilStruct
	}
	sendBuf := &bytes.Buffer{}
	enc := jute.NewBinaryEncoder(sendBuf)

	for _, generatedStruct := range generated {
		if generatedStruct == nil {
			return nil, errNilStruct
		}
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
