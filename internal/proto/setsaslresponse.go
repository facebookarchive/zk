// Autogenerated jute compiler
// @generated from '/Users/ardi/go/src/github.com/facebookincubator/zk/internal/zookeeper.jute'

package proto // github.com/facebookincubator/zk/internal/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type SetSASLResponse struct {
	Token []byte // token
}

func (r *SetSASLResponse) Read(dec jute.Decoder) (err error) {
	if err = dec.ReadStart(); err != nil {
		return err
	}
	r.Token, err = dec.ReadBuffer()
	if err != nil {
		return err
	}
	if err = dec.ReadEnd(); err != nil {
		return err
	}
	return nil
}

func (r *SetSASLResponse) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteBuffer(r.Token); err != nil {
		return err
	}
	if err := enc.WriteEnd(); err != nil {
		return err
	}
	return nil
}

func (r *SetSASLResponse) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("SetSASLResponse(%+v)", *r)
}
