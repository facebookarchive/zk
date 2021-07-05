package hostprovider

// HostProvider represents a set of hosts a ZooKeeper client should connect to.
type HostProvider interface {
	// Init is called first, with the servers specified in the connection string.
	Init(servers []string) error

	Size() int

	// Next returns the next server to connect to. retryStart will be true
	// if we've looped through all known servers without OnConnected being
	// called.
	Next() (server string, retryStart bool)

	// OnConnected notifies the HostProvider of a successful connection.
	// The HostProvider may use this notification to reset it's inner state.
	OnConnected()
}
