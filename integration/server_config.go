package integration

import (
	"fmt"
	"io"
	"strconv"
)

const (
	defaultPort                           = 2181
	defaultServerTickTime                 = 500
	defaultServerInitLimit                = 10
	defaultServerSyncLimit                = 5
	defaultServerAutoPurgeSnapRetainCount = 3
	defaultPeerPort                       = 2888
	defaultLeaderElectionPort             = 3888
)

// ServerConfigServer represents a single server in ServerConfig when replicated mode is used.
type ServerConfigServer struct {
	ID                 int
	Host               string
	PeerPort           int
	LeaderElectionPort int
}

// ServerConfig allows programmatic configurability of a Zookeeper server.
type ServerConfig struct {
	TickTime                 int    // Number of milliseconds of each tick
	InitLimit                int    // Number of ticks that the initial synchronization phase can take
	SyncLimit                int    // Number of ticks that can pass between sending a request and getting an acknowledgement
	DataDir                  string // Directory where the snapshot is stored
	ClientPort               int    // Port at which clients will connect
	AutoPurgeSnapRetainCount int    // Number of snapshots to retain in dataDir
	AutoPurgePurgeInterval   int    // Purge task internal in hours (0 to disable auto purge)
	FLWCommandsWhitelist     string
	Servers                  []*ServerConfigServer
	ReconfigEnabled          bool
}

// DefaultConfig returns a ServerConfig with default values.
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		TickTime:             defaultServerTickTime,
		InitLimit:            defaultServerInitLimit,
		SyncLimit:            defaultServerSyncLimit,
		DataDir:              "/tmp/gozk",
		ClientPort:           defaultPort,
		FLWCommandsWhitelist: "*",
		Servers:              nil,
	}
}

// Marshall writes the context of a ServerConfig struct to a file format understandable by Apache Zookeeper.
func (sc ServerConfig) Marshall(w io.Writer) error {
	// the admin server is not wanted in test cases as it slows the startup process and is
	// of little unit test value.
	fmt.Fprintln(w, "admin.enableServer=false")
	if sc.DataDir == "" {
		return fmt.Errorf("zk: missing dataDir server config field")
	}
	fmt.Fprintf(w, "dataDir=%s\n", sc.DataDir)
	if sc.TickTime <= 0 {
		sc.TickTime = defaultServerTickTime
	}
	fmt.Fprintf(w, "tickTime=%d\n", sc.TickTime)
	if sc.InitLimit <= 0 {
		sc.InitLimit = defaultServerInitLimit
	}
	fmt.Fprintf(w, "initLimit=%d\n", sc.InitLimit)
	if sc.SyncLimit <= 0 {
		sc.SyncLimit = defaultServerSyncLimit
	}
	fmt.Fprintf(w, "syncLimit=%d\n", sc.SyncLimit)
	if sc.ClientPort <= 0 {
		sc.ClientPort = defaultPort
	}
	fmt.Fprintf(w, "clientPort=%d\n", sc.ClientPort)
	if sc.AutoPurgePurgeInterval > 0 {
		if sc.AutoPurgeSnapRetainCount <= 0 {
			sc.AutoPurgeSnapRetainCount = defaultServerAutoPurgeSnapRetainCount
		}
		fmt.Fprintf(w, "autopurge.snapRetainCount=%d\n", sc.AutoPurgeSnapRetainCount)
		fmt.Fprintf(w, "autopurge.purgeInterval=%d\n", sc.AutoPurgePurgeInterval)
	}

	fmt.Fprintf(w, "reconfigEnabled=%s\n", strconv.FormatBool(sc.ReconfigEnabled))
	fmt.Fprintf(w, "4lw.commands.whitelist=%s\n", sc.FLWCommandsWhitelist)

	if len(sc.Servers) < 2 {
		// if we don't have more than 2 servers, we just don't specify server list to start in standalone mode
		// see https://zookeeper.apache.org/doc/current/zookeeperStarted.html#sc_InstallingSingleMode for more details.
		return nil
	}
	// if we then have more than one server, force it to be distributed
	fmt.Fprintln(w, "standaloneEnabled=false")

	for _, srv := range sc.Servers {
		if srv.PeerPort <= 0 {
			srv.PeerPort = defaultPeerPort
		}
		if srv.LeaderElectionPort <= 0 {
			srv.LeaderElectionPort = defaultLeaderElectionPort
		}
		fmt.Fprintf(w, "server.%d=%s:%d:%d\n", srv.ID, srv.Host, srv.PeerPort, srv.LeaderElectionPort)
	}
	return nil
}
