package zk

import "fmt"

// Error is an error code returned in a ReplyHeader by a Zookeeper server.
type Error int32

func (e Error) Error() string {
	if err, ok := errToString[e]; ok {
		return err
	}

	return fmt.Sprintf("unknown error code: %d", e)
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
	errAPI        Error = -100
	errNoNode     Error = -101
	errNoAuth     Error = -102
	errBadVersion Error = -103
	errNoChildren Error = -108
	errNodeExists Error = -110
	errNotEmpty   Error = -111
	errExpired    Error = -112
	errCallback   Error = -113
	errInvalidACL Error = -114
	errAuthFailed Error = -115
	errClosing    Error = -116
	errNothing    Error = -117
	errMoved      Error = -118
	errReadOnly   Error = -119
	errEphLocal   Error = -120
	errNoWatcher  Error = -121
	errReconfig   Error = -123
	errAuthScheme Error = -124
	errThrottled  Error = -127
)

var errToString = map[Error]string{
	errSys:           "system error",
	errRuntime:       "runtime inconsistency found",
	errData:          "data inconsistency found",
	errConnLoss:      "connection to the server has been lost",
	errMarshal:       "error while marshalling or unmarshalling data",
	errUnimpl:        "operation is unimplemented",
	errTimeout:       "operation timeout",
	errArgs:          "invalid arguments",
	errState:         "invalid zhandle state",
	errQuorum:        "no quorum of new config is connected",
	errCfgInProgress: "reconfiguration requested while another is currently in progress",
	errSSL:           "SSL connection error",

	errAPI:        "api error",
	errNoNode:     "node does not exist",
	errNoAuth:     "not authenticated",
	errBadVersion: "version conflict",
	errNoChildren: "ephemeral nodes may not have children",
	errNodeExists: "node already exists",
	errNotEmpty:   "node has children",
	errExpired:    "session has been expired by the server",
	errCallback:   "invalid callback specified",
	errInvalidACL: "invalid ACL specified",
	errAuthFailed: "client authentication failed",
	errClosing:    "zookeeper is closing",
	errNothing:    "no server responses to process",
	errMoved:      "session moved to another server, so operation is ignored",
	errReadOnly:   "state-changing request is passed to read-only server",
	errEphLocal:   "attempt to create ephemeral node on a local session",
	errNoWatcher:  "the watcher couldn't be found",
	errReconfig:   "attempts to perform a reconfiguration operation when it is disabled",
	errAuthScheme: "server requires configured authentication scheme",
	errThrottled:  "operation was throttled and not executed at all",
}
