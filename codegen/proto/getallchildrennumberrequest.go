// Autogenerated jute compiler
// @generated from 'zookeeper.jute'

package proto // github.com/facebookincubator/zk/codegen/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type GetAllChildrenNumberRequest struct {
	Path string // path
}

func (r *GetAllChildrenNumberRequest) GetPath() string {
	if r != nil {
		return r.Path
	}
	return ""
}

func (r *GetAllChildrenNumberRequest) Read(dec jute.Decoder) (err error) {
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

func (r *GetAllChildrenNumberRequest) Write(enc jute.Encoder) error {
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

func (r *GetAllChildrenNumberRequest) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("GetAllChildrenNumberRequest(%+v)", *r)
}
