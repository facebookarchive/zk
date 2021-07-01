package dnshostprovider

import (
	"testing"
)

// TestDNSHostProviderRetryStart tests the `retryStart` functionality
// of DNSHostProvider.
// It's also probably the clearest visual explanation of exactly how
// it works.
func TestDNSHostProviderRetryStart(t *testing.T) {
	t.Parallel()

	hp := &DNSHostProvider{lookupHostFunc: func(host string) ([]string, error) {
		return []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}, nil
	}}

	if err := hp.Init([]string{"foo.example.com:12345"}); err != nil {
		t.Fatal(err)
	}

	testdata := []struct {
		retryStartWant bool
		callConnected  bool
	}{
		// Repeated failures.
		{false, false},
		{false, false},
		{false, false},
		{true, false},
		{false, false},
		{false, false},
		{true, true},

		// One success offsets things.
		{false, false},
		{false, true},
		{false, true},

		// Repeated successes.
		{false, true},
		{false, true},
		{false, true},
		{false, true},
		{false, true},

		// And some more failures.
		{false, false},
		{false, false},
		{true, false}, // Looped back to last known good server: all alternates failed.
		{false, false},
	}

	for i, td := range testdata {
		_, retryStartGot := hp.Next()
		if retryStartGot != td.retryStartWant {
			t.Errorf("%d: retryStart=%v; want %v", i, retryStartGot, td.retryStartWant)
		}
		if td.callConnected {
			hp.OnConnected()
		}
	}
}
