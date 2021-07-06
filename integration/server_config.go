package integration

import (
	"fmt"
	"io"
	"strconv"
)

const (
	DefaultPort                           = 2181
	DefaultServerTickTime                 = 500
	DefaultServerInitLimit                = 10
	DefaultServerSyncLimit                = 5
	DefaultServerAutoPurgeSnapRetainCount = 3
	DefaultPeerPort                       = 2888
	DefaultLeaderElectionPort             = 3888
)

type ServerConfigServer struct {
	ID                 int
	Host               string
	PeerPort           int
	LeaderElectionPort int
}

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

type ErrMissingServerConfigField string

func (e ErrMissingServerConfigField) Error() string {
	return fmt.Sprintf("zk: missing server config field '%s'", string(e))
}

func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		TickTime:             500,
		InitLimit:            10,
		SyncLimit:            5,
		DataDir:              "/tmp/gozk",
		ClientPort:           2181,
		FLWCommandsWhitelist: "*",
		Servers:              nil,
	}
}

func (sc ServerConfig) Marshall(w io.Writer) error {
	// the admin server is not wanted in test cases as it slows the startup process and is
	// of little unit test value.
	fmt.Fprintln(w, "admin.enableServer=false")
	if sc.DataDir == "" {
		return ErrMissingServerConfigField("dataDir")
	}
	fmt.Fprintf(w, "dataDir=%s\n", sc.DataDir)
	if sc.TickTime <= 0 {
		sc.TickTime = DefaultServerTickTime
	}
	fmt.Fprintf(w, "tickTime=%d\n", sc.TickTime)
	if sc.InitLimit <= 0 {
		sc.InitLimit = DefaultServerInitLimit
	}
	fmt.Fprintf(w, "initLimit=%d\n", sc.InitLimit)
	if sc.SyncLimit <= 0 {
		sc.SyncLimit = DefaultServerSyncLimit
	}
	fmt.Fprintf(w, "syncLimit=%d\n", sc.SyncLimit)
	if sc.ClientPort <= 0 {
		sc.ClientPort = DefaultPort
	}
	fmt.Fprintf(w, "clientPort=%d\n", sc.ClientPort)
	if sc.AutoPurgePurgeInterval > 0 {
		if sc.AutoPurgeSnapRetainCount <= 0 {
			sc.AutoPurgeSnapRetainCount = DefaultServerAutoPurgeSnapRetainCount
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
			srv.PeerPort = DefaultPeerPort
		}
		if srv.LeaderElectionPort <= 0 {
			srv.LeaderElectionPort = DefaultLeaderElectionPort
		}
		fmt.Fprintf(w, "server.%d=%s:%d:%d\n", srv.ID, srv.Host, srv.PeerPort, srv.LeaderElectionPort)
	}
	return nil
}
