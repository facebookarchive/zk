// Autogenerated jute compiler
// @generated from '/Users/ardi/go/src/github.com/facebookincubator/zk/internal/zookeeper.jute'

package proto // github.com/facebookincubator/zk/internal/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type GetEphemeralsRequest struct {
	PrefixPath *string // prefixPath
}

func (r *GetEphemeralsRequest) Read(dec jute.Decoder) (err error) {
	if err = dec.ReadStart(); err != nil {
		return err
	}
	r.PrefixPath, err = dec.ReadUstring()
	if err != nil {
		return err
	}
	if err = dec.ReadEnd(); err != nil {
		return err
	}
	return nil
}

func (r *GetEphemeralsRequest) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteUstring(r.PrefixPath); err != nil {
		return err
	}
	if err := enc.WriteEnd(); err != nil {
		return err
	}
	return nil
}

func (r *GetEphemeralsRequest) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("GetEphemeralsRequest(%+v)", *r)
}
