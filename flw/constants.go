package flw

import "time"

// defaultPort is the default port number on which Zookeeper servers listen for client connections.
const defaultPort = 2181
const defaultTimeout = 2 * time.Second

// Mode is used to build custom server modes (leader|follower|standalone).
type Mode uint8

func (m Mode) String() string {
	if name := modeNames[m]; name != "" {
		return name
	}
	return "unknown"
}

// These constants represent the possible modes that a Zookeeper server can run in. A Zookeeper server can
// run in leader, follower or standalone mode. ModeUnknown is used when the mode cannot be recognized from the response.
const (
	ModeUnknown Mode = iota
	ModeLeader
	ModeFollower
	ModeStandalone
)

var (
	modeNames = map[Mode]string{
		ModeLeader:     "leader",
		ModeFollower:   "follower",
		ModeStandalone: "standalone",
	}
)
