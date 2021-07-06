// Autogenerated jute compiler
// @generated from 'zookeeper.jute'

package proto // github.com/facebookincubator/zk/codegen/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type ReplyHeader struct {
	Xid  int32 // xid
	Zxid int64 // zxid
	Err  int32 // err
}

func (r *ReplyHeader) GetXid() int32 {
	if r != nil {
		return r.Xid
	}
	return 0
}

func (r *ReplyHeader) GetZxid() int64 {
	if r != nil {
		return r.Zxid
	}
	return 0
}

func (r *ReplyHeader) GetErr() int32 {
	if r != nil {
		return r.Err
	}
	return 0
}

func (r *ReplyHeader) Read(dec jute.Decoder) (err error) {
	if err = dec.ReadStart(); err != nil {
		return err
	}
	r.Xid, err = dec.ReadInt()
	if err != nil {
		return err
	}
	r.Zxid, err = dec.ReadLong()
	if err != nil {
		return err
	}
	r.Err, err = dec.ReadInt()
	if err != nil {
		return err
	}
	if err = dec.ReadEnd(); err != nil {
		return err
	}
	return nil
}

func (r *ReplyHeader) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteInt(r.Xid); err != nil {
		return err
	}
	if err := enc.WriteLong(r.Zxid); err != nil {
		return err
	}
	if err := enc.WriteInt(r.Err); err != nil {
		return err
	}
	if err := enc.WriteEnd(); err != nil {
		return err
	}
	return nil
}

func (r *ReplyHeader) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("ReplyHeader(%+v)", *r)
}
