package flw

// DefaultPort is the default port number on which Zookeeper servers listen for client connections.
const DefaultPort = 2181

// Mode is used to build custom server modes (leader|follower|standalone).
type Mode uint8

func (m Mode) String() string {
	if name := modeNames[m]; name != "" {
		return name
	}
	return "unknown"
}

const (
	modeUnknown    Mode = iota
	modeLeader     Mode = iota
	modeFollower   Mode = iota
	modeStandalone Mode = iota
)

var (
	modeNames = map[Mode]string{
		modeLeader:     "leader",
		modeFollower:   "follower",
		modeStandalone: "standalone",
	}
)
