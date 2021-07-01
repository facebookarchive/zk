package dnshostprovider

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
)

// DNSHostProvider is the default HostProvider. It currently matches
// the Java StaticHostProvider, resolving hosts from DNS once during
// the call to Init.  It could be easily extended to re-query DNS
// periodically or if there is trouble connecting.
type DNSHostProvider struct {
	mu             sync.Mutex
	servers        []string
	curr           int
	last           int
	lookupHostFunc func(string) ([]string, error) // Override of net.LookupHost, for testing.
}

// Init is called first, with the servers specified in the connection
// string. It uses DNS to look up addresses for each server, then
// shuffles them all together.
func (hp *DNSHostProvider) Init(servers []string) error {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	lookupHostFunc := hp.lookupHostFunc
	if lookupHostFunc == nil {
		lookupHostFunc = net.LookupHost
	}

	var found []string
	for _, server := range servers {
		host, port, err := net.SplitHostPort(server)
		if err != nil {
			return err
		}
		addrs, err := lookupHostFunc(host)
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			found = append(found, net.JoinHostPort(addr, port))
		}
	}

	if len(found) == 0 {
		return fmt.Errorf("No hosts found for addresses %q", servers)
	}

	// Randomize the order of the servers to avoid creating hotspots
	stringShuffle(found)

	hp.servers = found
	hp.curr = -1
	hp.last = -1

	return nil
}

// Size returns the number of servers available
func (hp *DNSHostProvider) Size() int {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	return len(hp.servers)
}

// Next returns the next server to connect to. retryStart will be true
// if we've looped through all known servers without OnConnected() being
// called.
func (hp *DNSHostProvider) Next() (server string, retryStart bool) {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	hp.curr = (hp.curr + 1) % len(hp.servers)
	retryStart = hp.curr == hp.last
	if hp.last == -1 {
		hp.last = 0
	}
	return hp.servers[hp.curr], retryStart
}

// OnConnected notifies the HostProvider of a successful connection.
func (hp *DNSHostProvider) OnConnected() {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	hp.last = hp.curr
}

// stringShuffle performs a Fisher-Yates shuffle on a slice of strings
func stringShuffle(s []string) {
	for i := len(s) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}
