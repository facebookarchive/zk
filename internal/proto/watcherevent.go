// Autogenerated jute compiler
// @generated from 'zookeeper.jute'

package proto // github.com/facebookincubator/zk/internal/proto

import (
	"fmt"

	"github.com/go-zookeeper/jute/lib/go/jute"
)

type WatcherEvent struct {
	Type  int32  // type
	State int32  // state
	Path  string // path
}

func (r *WatcherEvent) GetType() int32 {
	if r != nil {
		return r.Type
	}
	return 0
}

func (r *WatcherEvent) GetState() int32 {
	if r != nil {
		return r.State
	}
	return 0
}

func (r *WatcherEvent) GetPath() string {
	if r != nil {
		return r.Path
	}
	return ""
}

func (r *WatcherEvent) Read(dec jute.Decoder) (err error) {
	if err = dec.ReadStart(); err != nil {
		return err
	}
	r.Type, err = dec.ReadInt()
	if err != nil {
		return err
	}
	r.State, err = dec.ReadInt()
	if err != nil {
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

func (r *WatcherEvent) Write(enc jute.Encoder) error {
	if err := enc.WriteStart(); err != nil {
		return err
	}
	if err := enc.WriteInt(r.Type); err != nil {
		return err
	}
	if err := enc.WriteInt(r.State); err != nil {
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

func (r *WatcherEvent) String() string {
	if r == nil {
		return "<nil>"
	}
	return fmt.Sprintf("WatcherEvent(%+v)", *r)
}