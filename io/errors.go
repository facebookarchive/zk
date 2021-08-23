package io

// Error is a wrapper for all error codes that can be returned by a Zookeeper server.
type Error struct {
	Code Code
}

// Code is an error code returned in a ReplyHeader by a Zookeeper server.
type Code int32

func (e Error) Error() string {
	if err, ok := errToString[e.Code]; ok {
		return err
	}

	return ""
}

// ref: https://github.com/apache/zookeeper/blob/master/zookeeper-client/zookeeper-client-c/include/zookeeper.h#L94
const (
	// System and server-side errors
	errSys           = -1
	errRuntime       = -2
	errData          = -3
	errConnLoss      = -4
	errMarshal       = -5
	errUnimpl        = -6
	errTimeout       = -7
	errArgs          = -8
	errState         = -9
	errQuorum        = -13
	errCfgInProgress = -14
	errSSL           = -15

	// API errors
	errAPI        Code = -100
	errNoNode     Code = -101
	errNoAuth     Code = -102
	errBadVersion Code = -103
	errNoChildren Code = -108
	errNodeExists Code = -110
	errNotEmpty   Code = -111
	errExpired    Code = -112
	errCallback   Code = -113
	errInvalidACL Code = -114
	errAuthFailed Code = -115
	errClosing    Code = -116
	errNothing    Code = -117
	errMoved      Code = -118
	errReadOnly   Code = -119
	errEphLocal   Code = -120
	errNoWatcher  Code = -121
	errReconfig   Code = -123
	errAuthScheme Code = -124
	errThrottled  Code = -127
)

var errToString = map[Code]string{
	errSys:           "zk: system error",
	errRuntime:       "zk: runtime inconsistency found",
	errData:          "zk: data inconsistency found",
	errConnLoss:      "zk: connection to the server has been lost",
	errMarshal:       "zk: error while marshalling or unmarshalling data",
	errUnimpl:        "zk: operation is unimplemented",
	errTimeout:       "zk: operation timeout",
	errArgs:          "zk: invalid arguments",
	errState:         "zk: invalid zhandle state",
	errQuorum:        "zk: no quorum of new config is connected",
	errCfgInProgress: "zk: reconfiguration requested while another is currently in progress",
	errSSL:           "zk: SSL connection error",

	errAPI:        "zk: api error",
	errNoNode:     "zk: node does not exist",
	errNoAuth:     "zk: not authenticated",
	errBadVersion: "zk: version conflict",
	errNoChildren: "zk: ephemeral nodes may not have children",
	errNodeExists: "zk: node already exists",
	errNotEmpty:   "zk: node has children",
	errExpired:    "zk: session has been expired by the server",
	errCallback:   "zk: invalid callback specified",
	errInvalidACL: "zk: invalid ACL specified",
	errAuthFailed: "zk: client authentication failed",
	errClosing:    "zk: zookeeper is closing",
	errNothing:    "zk: no server responses to process",
	errMoved:      "zk: session moved to another server, so operation is ignored",
	errReadOnly:   "zk: state-changing request is passed to read-only server",
	errEphLocal:   "zk: attempt to create ephemeral node on a local session",
	errNoWatcher:  "zk: the watcher couldn't be found",
	errReconfig:   "zk: attempts to perform a reconfiguration operation when it is disabled",
	errAuthScheme: "zk: server requires configured authentication scheme",
	errThrottled:  "zk: operation was throttled and not executed at all",
}
