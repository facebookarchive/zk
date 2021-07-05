// Autogenerated jute compiler
// @generated from '/Users/ardi/go/src/github.com/facebookincubator/zk/internal/zookeeper.jute'

package proto // github.com/facebookincubator/zk/internal/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type CheckVersionRequest struct {
	Path    *string // path
	Version int32   // version
}

func (r *CheckVersionRequest) Read(dec jute.Decoder) (err error) {
	if err = dec.ReadStart(); err != nil {
		return err
	}
	r.Path, err = dec.ReadUstring()
	if err != nil {
		return err
	}
	r.Version, err = dec.ReadInt()
	if err != nil {
		return err
	}
	if err = dec.ReadEnd(); err != nil {
		return err
	}
	return nil
}

func (r *CheckVersionRequest) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteUstring(r.Path); err != nil {
		return err
	}
	if err := enc.WriteInt(r.Version); err != nil {
		return err
	}
	if err := enc.WriteEnd(); err != nil {
		return err
	}
	return nil
}

func (r *CheckVersionRequest) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("CheckVersionRequest(%+v)", *r)
}
