package flw

import "time"

// DefaultPort is the default port number on which Zookeeper servers listen for client connections.
const DefaultPort = 2181
const defaultTimeout = 2 * time.Second

// Mode is used to build custom server modes (leader|follower|standalone).
type Mode uint8

func (m Mode) String() string {
	if name := modeNames[m]; name != "" {
		return name
	}
	return "unknown"
}

const (
	modeUnknown Mode = iota
	modeLeader
	modeFollower
	modeStandalone
)

var (
	modeNames = map[Mode]string{
		modeLeader:     "leader",
		modeFollower:   "follower",
		modeStandalone: "standalone",
	}
)
