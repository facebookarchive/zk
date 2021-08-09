package flw

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Srvr is a FourLetterWord helper function. In particular, this function pulls the "srvr" output
// from the zookeeper instance and parses the output. A *ServerStats struct is returned
// as well as an error value to indicate whether this function processed successfully.
func (c *Client) Srvr(server string) (*ServerStats, error) {
	// different parts of the regular expression that are required to parse the srvr output
	const (
		zrVer   = `^Zookeeper version: ([A-Za-z0-9\.\-]+), built on (\d\d/\d\d/\d\d\d\d \d\d:\d\d [A-Za-z0-9:\+\-]+)`
		zrLat   = `^Latency min/avg/max: (\d+)/([0-9.]+)/(\d+)`
		zrNet   = `^Received: (\d+).*\n^Sent: (\d+).*\n^Connections: (\d+).*\n^Outstanding: (\d+)`
		zrState = `^Zxid: (0x[A-Za-z0-9]+).*\n^Mode: (\w+).*\n^Node count: (\d+)`
	)

	// build the regex from the pieces above
	re, err := regexp.Compile(fmt.Sprintf(`(?m:\A%v.*\n%v.*\n%v.*\n%v)`, zrVer, zrLat, zrNet, zrState))
	if err != nil {
		return nil, fmt.Errorf("error compiling srvr response regex: %w", err)
	}

	response, err := fourLetterWord(server, "srvr", c.Timeout)

	if err != nil {
		return nil, fmt.Errorf("invalid srvr response: %w", err)
	}
	matches := re.FindAllStringSubmatch(string(response), -1)

	if matches == nil {
		return nil, fmt.Errorf("unable to parse fields from zookeeper response (no regex matches)")
	}

	match := matches[0][1:]

	// determine current server
	var srvrMode Mode
	switch match[10] {
	case "leader":
		srvrMode = ModeLeader
	case "follower":
		srvrMode = ModeFollower
	case "standalone":
		srvrMode = ModeStandalone
	default:
		srvrMode = ModeUnknown
	}

	buildTime, err := time.Parse("01/02/2006 15:04 MST", match[1])

	if err != nil {
		return nil, fmt.Errorf("error parsing srvr response: %w", err)
	}

	parsedInt, err := strconv.ParseInt(match[9], 0, 64)

	if err != nil {
		return nil, fmt.Errorf("error parsing srvr response: %w", err)
	}

	// the ZxID value is an int64 with two int32s packed inside
	// the high int32 is the epoch (i.e., number of leader elections)
	// the low int32 is the counter
	epoch := int32(parsedInt >> 32)
	counter := int32(parsedInt & 0xFFFFFFFF)

	// within the regex above, these values must be numerical
	// so we can avoid useless checking of the error return value
	minLatency, _ := strconv.ParseInt(match[2], 0, 64)
	avgLatency, _ := strconv.ParseFloat(match[3], 64)
	maxLatency, _ := strconv.ParseInt(match[4], 0, 64)
	recv, _ := strconv.ParseInt(match[5], 0, 64)
	sent, _ := strconv.ParseInt(match[6], 0, 64)
	cons, _ := strconv.ParseInt(match[7], 0, 64)
	outs, _ := strconv.ParseInt(match[8], 0, 64)
	ncnt, _ := strconv.ParseInt(match[11], 0, 64)

	return &ServerStats{
		Sent:        sent,
		Received:    recv,
		NodeCount:   ncnt,
		MinLatency:  minLatency,
		AvgLatency:  avgLatency,
		MaxLatency:  maxLatency,
		Connections: cons,
		Outstanding: outs,
		Epoch:       epoch,
		Counter:     counter,
		BuildTime:   buildTime,
		Mode:        srvrMode,
		Version:     match[0],
	}, nil
}

// Ruok is a FourLetterWord helper function. In particular, this function
// pulls the "ruok" output of a server.
func (c *Client) Ruok(server string) error {
	response, err := fourLetterWord(server, "ruok", c.Timeout)
	if err != nil {
		return fmt.Errorf("error calling ruok FLW: %w", err)
	}

	if string(response[:4]) != "imok" {
		return fmt.Errorf("invalid ruok response from server: %s", response[:4])
	}

	return nil
}

// Cons is a FourLetterWord helper function. In particular, this function
// pulls the "cons" output from a server.
// As with Srvr, the error value indicates whether the request had an issue.
func (c *Client) Cons(server string) ([]*ServerClient, error) {
	const (
		zrAddr = `^ /((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?):(?:\d+))\[\d+\]`
		zrPac  = `\(queued=(\d+),recved=(\d+),sent=(\d+),sid=(0x[A-Za-z0-9]+),lop=(\w+),est=(\d+),to=(\d+),`
		zrSesh = `lcxid=(0x[A-Za-z0-9]+),lzxid=(0x[A-Za-z0-9]+),lresp=(\d+),llat=(\d+),minlat=(\d+),avglat=(\d+),maxlat=(\d+)\)`
	)

	re, err := regexp.Compile(fmt.Sprintf("%v%v%v", zrAddr, zrPac, zrSesh))
	if err != nil {
		return nil, fmt.Errorf("error compiling cons response regex: %w", err)
	}

	response, err := fourLetterWord(server, "cons", c.Timeout)

	if err != nil {
		return nil, fmt.Errorf("error parsing cons response: %w", err)
	}

	scan := bufio.NewScanner(bytes.NewReader(response))

	var clients []*ServerClient

	for scan.Scan() {
		line := scan.Bytes()

		if len(line) == 0 {
			continue
		}

		m := re.FindAllStringSubmatch(string(line), -1)

		if m == nil {
			return nil, fmt.Errorf("unable to parse fields from zookeeper response (no regex matches)")
		}

		match := m[0][1:]

		queued, _ := strconv.ParseInt(match[1], 0, 64)
		recvd, _ := strconv.ParseInt(match[2], 0, 64)
		sent, _ := strconv.ParseInt(match[3], 0, 64)
		sid, _ := strconv.ParseInt(match[4], 0, 64)
		est, _ := strconv.ParseInt(match[6], 0, 64)
		timeout, _ := strconv.ParseInt(match[7], 0, 32)
		lcxid, _ := parseInt64(match[8])
		lzxid, _ := parseInt64(match[9])
		lresp, _ := strconv.ParseInt(match[10], 0, 64)
		llat, _ := strconv.ParseInt(match[11], 0, 32)
		minlat, _ := strconv.ParseInt(match[12], 0, 32)
		avglat, _ := strconv.ParseInt(match[13], 0, 32)
		maxlat, _ := strconv.ParseInt(match[14], 0, 32)

		clients = append(clients, &ServerClient{
			Queued:        queued,
			Received:      recvd,
			Sent:          sent,
			SessionID:     sid,
			Lcxid:         int64(lcxid),
			Lzxid:         int64(lzxid),
			Timeout:       int32(timeout),
			LastLatency:   int32(llat),
			MinLatency:    int32(minlat),
			AvgLatency:    int32(avglat),
			MaxLatency:    int32(maxlat),
			Established:   time.Unix(est, 0),
			LastResponse:  time.Unix(lresp, 0),
			Addr:          match[0],
			LastOperation: match[5],
		})
	}

	return clients, nil
}

// Srvr executes the srvr FLW protocol function using the default client.
func Srvr(server string) (*ServerStats, error) {
	defaultClient := &Client{Timeout: defaultTimeout}
	return defaultClient.Srvr(server)
}

// Ruok executes the ruok FLW protocol function using the default client.
func Ruok(server string) error {
	defaultClient := &Client{Timeout: defaultTimeout}
	return defaultClient.Ruok(server)
}

// Cons executes the cons FLW protocol function using the default client.
func Cons(server string) ([]*ServerClient, error) {
	defaultClient := &Client{Timeout: defaultTimeout}
	return defaultClient.Cons(server)
}

// parseInt64 is similar to strconv.ParseInt, but it also handles hex values that represent negative numbers
func parseInt64(s string) (int64, error) {
	if strings.HasPrefix(s, "0x") {
		i, err := strconv.ParseUint(s, 0, 64)
		return int64(i), err
	}
	return strconv.ParseInt(s, 0, 64)
}

func fourLetterWord(server, command string, timeout time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", server, timeout)
	if err != nil {
		return nil, err
	}

	// the zookeeper server should automatically close this socket
	// once the command has been processed, but better safe than sorry
	defer conn.Close()

	if err = conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	_, err = conn.Write([]byte(command))
	if err != nil {
		return nil, err
	}

	if err = conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(conn)
}
