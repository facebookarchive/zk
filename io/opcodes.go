package io

// PingXID represents the XID which is used in ping/keepalive packet headers.
const PingXID = -2

// Below constants represent codes used by Zookeeper to differentiate requests.
// https://zookeeper.apache.org/doc/r3.4.8/api/constant-values.html#org.apache.zookeeper.ZooDefs.OpCode.getData
const (
	OpGetData     = 4
	OpGetChildren = 8
	OpPing        = 11
)

type Code int32
type Error struct {
	Code Code
}

// ref: https://github.com/apache/zookeeper/blob/master/zookeeper-client/zookeeper-client-c/include/zookeeper.h#L94
const (
	// System and server-side errors
	errSys       = -1
	errRuntime   = -2
	errData      = -3
	errConnLoss  = -4
	errMarshal   = -5
	errUnimpl    = -6
	errOpTimeout = -7
	errArgs      = -8
	errState     = -9

	// API errors
	errAPI        Code = -100
	errNoNode     Code = -101 // *
	errNoAuth     Code = -102
	errBadVersion Code = -103 // *
	errNoChildren Code = -108
	errNodeExists Code = -110 // *
	errNotEmpty   Code = -111
	errExpired    Code = -112
	errCallback   Code = -113
	errInvalidAcl Code = -114
	errAuthFailed Code = -115
	errClosing    Code = -116
	errNothing    Code = -117
	errMoved      Code = -118
	// Attempts to perform a reconfiguration operation when reconfiguration feature is disabled
	errZReconfigDisabled Code = -123
)

var errToString = map[Code]string{
	errAPI:               "zk: api error",
	errNoNode:            "zk: node does not exist",
	errNoAuth:            "zk: not authenticated",
	errBadVersion:        "zk: version conflict",
	errNoChildren:        "zk: ephemeral nodes may not have children",
	errNodeExists:        "zk: node already exists",
	errNotEmpty:          "zk: node has children",
	errExpired:           "zk: session has been expired by the server",
	errCallback:          "zk: invalid callback specified",
	errInvalidAcl:        "zk: invalid ACL specified",
	errAuthFailed:        "zk: client authentication failed",
	errClosing:           "zk: zookeeper is closing",
	errNothing:           "zk: no server responses to process",
	errMoved:             "zk: session moved to another server, so operation is ignored",
	errZReconfigDisabled: "zk: attempts to perform a reconfiguration operation when it is disabled",
	errArgs:              "zk: invalid arguments",
}

func (e Error) Error() string {
	if err, ok := errToString[e.Code]; ok {
		return err
	}

	return ""
}
