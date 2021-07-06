// Autogenerated jute compiler
// @generated from 'zookeeper.jute'

package proto // github.com/facebookincubator/zk/codegen/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type GetMaxChildrenRequest struct {
	Path string // path
}

func (r *GetMaxChildrenRequest) GetPath() string {
	if r != nil {
		return r.Path
	}
	return ""
}

func (r *GetMaxChildrenRequest) Read(dec jute.Decoder) (err error) {
	if err = dec.ReadStart(); err != nil {
		return err
	}
	r.Path, err = dec.ReadString()
	if err != nil {
		return err
	}
	if err = dec.ReadEnd(); err != nil {
		return err
	}
	return nil
}

func (r *GetMaxChildrenRequest) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteString(r.Path); err != nil {
		return err
	}
	if err := enc.WriteEnd(); err != nil {
		return err
	}
	return nil
}

func (r *GetMaxChildrenRequest) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("GetMaxChildrenRequest(%+v)", *r)
}
