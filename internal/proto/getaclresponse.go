// Autogenerated jute compiler
// @generated from '/Users/ardi/go/src/github.com/facebookincubator/zk/internal/zookeeper.jute'

package proto // github.com/facebookincubator/zk/internal/proto

import (
	"fmt"

	"github.com/facebookincubator/zk/internal/data"
	"github.com/go-zookeeper/jute/lib/go/jute"
)

type GetACLResponse struct {
	Acl  []*data.ACL // acl
	Stat *data.Stat  // stat
}

func (r *GetACLResponse) Read(dec jute.Decoder) (err error) {
	var size int
	if err = dec.ReadStart(); err != nil {
		return err
	}
	size, err = dec.ReadVectorStart()
	if err != nil {
		return err
	}
	if size < 0 {
		r.Acl = nil
	} else {
		r.Acl = make([]*data.ACL, size)
		for i := 0; i < size; i++ {
			if err = dec.ReadRecord(r.Acl[i]); err != nil {
				return err
			}
		}
	}
	if err = dec.ReadVectorEnd(); err != nil {
		return err
	}
	if err = dec.ReadRecord(r.Stat); err != nil {
		return err
	}
	if err = dec.ReadEnd(); err != nil {
		return err
	}
	return nil
}

func (r *GetACLResponse) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteVectorStart(len(r.Acl), r.Acl == nil); err != nil {
		return err
	}
	for _, v := range r.Acl {
		if err := enc.WriteRecord(v); err != nil {
			return err
		}
	}
	if err := enc.WriteVectorEnd(); err != nil {
		return err
	}
	if err := enc.WriteRecord(r.Stat); err != nil {
		return err
	}
	if err := enc.WriteEnd(); err != nil {
		return err
	}
	return nil
}

func (r *GetACLResponse) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("GetACLResponse(%+v)", *r)
}
