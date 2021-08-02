package flw

import (
	"net"
	"testing"
	"time"
)

const (
	zkSrvrOut = `Zookeeper version: 3.4.6-1569965, built on 02/20/2014 09:09 GMT
Latency min/avg/max: 0/1/10
Received: 4207
Sent: 4220
Connections: 81
Outstanding: 1
Zxid: 0x110a7a8f37
Mode: leader
Node count: 306
`
	zkConsOut = ` /10.42.45.231:45361[1](queued=0,recved=9435,sent=9457,sid=0x94c2989e04716b5,lop=PING,est=1427238717217,to=20001,lcxid=0x55120915,lzxid=0xffffffffffffffff,lresp=1427259255908,llat=0,minlat=0,avglat=1,maxlat=17)
 /10.55.33.98:34342[1](queued=0,recved=9338,sent=9350,sid=0x94c2989e0471731,lop=PING,est=1427238849319,to=20001,lcxid=0x55120944,lzxid=0xffffffffffffffff,lresp=1427259252294,llat=0,minlat=0,avglat=1,maxlat=18)
 /10.44.145.114:46556[1](queued=0,recved=109253,sent=109617,sid=0x94c2989e0471709,lop=DELE,est=1427238791305,to=20001,lcxid=0x55139618,lzxid=0x110a7b187d,lresp=1427259257423,llat=2,minlat=0,avglat=1,maxlat=23)

`
)

func TestRuok(t *testing.T) {
	t.Parallel()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go tcpServer(l, "")

	client := &Client{timeout: 10 * time.Second}
	oks := client.Ruok([]string{l.Addr().String()})
	if len(oks) == 0 {
		t.Errorf("No values returned")
	}
	if !oks[0] {
		t.Errorf("Instance should be marked as OK")
	}

	//
	// Confirm that it also returns false for dead instances
	//
	l, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go tcpServer(l, "dead")

	oks = client.Ruok([]string{l.Addr().String()})
	if len(oks) == 0 {
		t.Errorf("No values returned")
	}
	if oks[0] {
		t.Errorf("Instance should be marked as not OK")
	}
}

func TestSrvr(t *testing.T) {
	t.Parallel()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go tcpServer(l, "")

	client := &Client{timeout: 10 * time.Second}
	statsSlice, ok := client.Srvr([]string{l.Addr().String()})
	if !ok {
		t.Errorf("Failure indicated on 'srvr' parsing")
	}
	if len(statsSlice) == 0 {
		t.Errorf("No *ServerStats instances returned")
	}
	expected := &ServerStats{
		Sent:        4220,
		Received:    4207,
		NodeCount:   306,
		MinLatency:  0,
		AvgLatency:  1,
		MaxLatency:  10,
		Connections: 81,
		Outstanding: 1,
		Epoch:       17,
		Counter:     175804215,
		Mode:        ModeLeader,
		Version:     "3.4.6-1569965",
	}
	stats := statsSlice[0]

	if stats.Error != nil {
		t.Fatalf("Unexpected error seen in stats: %v", err.Error())
	}

	if stats.Sent != expected.Sent {
		t.Errorf("Sent value mismatch (expected %d, got %d)", stats.Sent, expected.Sent)
	}

	if stats.Received != expected.Received {
		t.Errorf("Received value mismatch (expected %d, got %d)", stats.Received, expected.Received)
	}

	if stats.NodeCount != expected.NodeCount {
		t.Errorf("NodeCount value mismatch (expected %d, got %d)", stats.NodeCount, expected.NodeCount)
	}

	if stats.MinLatency != expected.MinLatency {
		t.Errorf("MinLatency value mismatch (expected %d, got %d)", stats.MinLatency, expected.MinLatency)
	}

	if stats.AvgLatency != expected.AvgLatency {
		t.Errorf("AvgLatency value mismatch (expected %f, got %f)", stats.AvgLatency, expected.AvgLatency)
	}

	if stats.MaxLatency != expected.MaxLatency {
		t.Errorf("MaxLatency value mismatch (expected %d, got %d)", stats.MaxLatency, expected.MaxLatency)
	}

	if stats.Connections != expected.Connections {
		t.Errorf("Connection value mismatch (expected %d, got %d)", stats.Connections, expected.Connections)
	}

	if stats.Outstanding != expected.Outstanding {
		t.Errorf("Outstanding value mismatch (expected %d, got %d)", stats.Outstanding, expected.Outstanding)
	}

	if stats.Epoch != expected.Epoch {
		t.Errorf("Epoch value mismatch (expected %d, got %d)", stats.Epoch, expected.Epoch)
	}

	if stats.Counter != expected.Counter {
		t.Errorf("Counter value mismatch (expected %d, got %d)", stats.Counter, expected.Counter)
	}

	if stats.Mode != expected.Mode {
		t.Errorf("Mode value mismatch (expected %v, got %v)", stats.Mode, expected.Mode)
	}

	if stats.Version != expected.Version {
		t.Errorf("Mode value mismatch (expected %v, got %v)", stats.Version, expected.Version)
	}
}

func TestCons(t *testing.T) {
	t.Parallel()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go tcpServer(l, "")

	client := &Client{timeout: 10 * time.Second}
	clients, ok := client.Cons([]string{l.Addr().String()})
	if !ok {
		t.Errorf("failure indicated on 'cons' parsing")
	}
	if len(clients) == 0 {
		t.Errorf("no *ServerClients instances returned")
	}

	results := []*ServerClient{
		{
			Queued:        0,
			Received:      9435,
			Sent:          9457,
			SessionID:     669956116721374901,
			LastOperation: "PING",
			Established:   time.Unix(1427238717217, 0),
			Timeout:       20001,
			Lcxid:         1427245333,
			Lzxid:         -1,
			LastResponse:  time.Unix(1427259255908, 0),
			LastLatency:   0,
			MinLatency:    0,
			AvgLatency:    1,
			MaxLatency:    17,
			Addr:          "10.42.45.231:45361",
		},
		{
			Queued:        0,
			Received:      9338,
			Sent:          9350,
			SessionID:     669956116721375025,
			LastOperation: "PING",
			Established:   time.Unix(1427238849319, 0),
			Timeout:       20001,
			Lcxid:         1427245380,
			Lzxid:         -1,
			LastResponse:  time.Unix(1427259252294, 0),
			LastLatency:   0,
			MinLatency:    0,
			AvgLatency:    1,
			MaxLatency:    18,
			Addr:          "10.55.33.98:34342",
		},
		{
			Queued:        0,
			Received:      109253,
			Sent:          109617,
			SessionID:     669956116721374985,
			LastOperation: "DELE",
			Established:   time.Unix(1427238791305, 0),
			Timeout:       20001,
			Lcxid:         1427346968,
			Lzxid:         73190283389,
			LastResponse:  time.Unix(1427259257423, 0),
			LastLatency:   2,
			MinLatency:    0,
			AvgLatency:    1,
			MaxLatency:    23,
			Addr:          "10.44.145.114:46556",
		},
	}

	for _, z := range clients {
		if z.Error != nil {
			t.Errorf("error seen: %v", err.Error())
		}

		for i, v := range z.Clients {
			c := results[i]

			if v.Error != nil {
				t.Errorf("Unexpected client error: %s", err.Error())
			}

			if v.Queued != c.Queued {
				t.Errorf("Queued value mismatch (expected %d, got %d)", v.Queued, c.Queued)
			}

			if v.Received != c.Received {
				t.Errorf("Received value mismatch (expected %d, got %d)", v.Received, c.Received)
			}

			if v.Sent != c.Sent {
				t.Errorf("Sent value mismatch (expected %d, got %d)", v.Sent, c.Sent)
			}

			if v.SessionID != c.SessionID {
				t.Errorf("SessionID value mismatch (expected %d, got %d)", v.SessionID, c.SessionID)
			}

			if v.LastOperation != c.LastOperation {
				t.Errorf("LastOperation value mismatch (expected %v, got %v)", v.LastOperation, c.LastOperation)
			}

			if v.Timeout != c.Timeout {
				t.Errorf("Timeout value mismatch (expected %d, got %d)", v.Timeout, c.Timeout)
			}

			if v.Lcxid != c.Lcxid {
				t.Errorf("Lcxid value mismatch (expected %d, got %d)", v.Lcxid, c.Lcxid)
			}

			if v.Lzxid != c.Lzxid {
				t.Errorf("Lzxid value mismatch (expected %d, got %d)", v.Lzxid, c.Lzxid)
			}

			if v.LastLatency != c.LastLatency {
				t.Errorf("LastLatency value mismatch (expected %d, got %d)", v.LastLatency, c.LastLatency)
			}

			if v.MinLatency != c.MinLatency {
				t.Errorf("MinLatency value mismatch (expected %d, got %d)", v.MinLatency, c.MinLatency)
			}

			if v.AvgLatency != c.AvgLatency {
				t.Errorf("AvgLatency value mismatch (expected %d, got %d)", v.AvgLatency, c.AvgLatency)
			}

			if v.MaxLatency != c.MaxLatency {
				t.Errorf("MaxLatency value mismatch (expected %d, got %d)", v.MaxLatency, c.MaxLatency)
			}

			if v.Addr != c.Addr {
				t.Errorf("Addr value mismatch (expected %v, got %v)", v.Addr, c.Addr)
			}

			if !c.Established.Equal(v.Established) {
				t.Errorf("Established value mismatch (expected %v, got %v)", c.Established, v.Established)
			}

			if !c.LastResponse.Equal(v.LastResponse) {
				t.Errorf("Established value mismatch (expected %v, got %v)", c.LastResponse, v.LastResponse)
			}
		}
	}
}

func tcpServer(listener net.Listener, status string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go connHandler(conn, status)
	}
}

func connHandler(conn net.Conn, status string) {
	defer conn.Close()

	data := make([]byte, 4)

	_, err := conn.Read(data)
	if err != nil {
		return
	}

	switch string(data) {
	case "ruok":
		switch status {
		case "dead":
			return
		default:
			conn.Write([]byte("imok"))
		}
	case "srvr":
		switch status {
		case "dead":
			return
		default:
			conn.Write([]byte(zkSrvrOut))
		}
	case "cons":
		switch status {
		case "dead":
			return
		default:
			conn.Write([]byte(zkConsOut))
		}
	default:
		conn.Write([]byte("This ZooKeeper instance is not currently serving requests."))
	}
}
