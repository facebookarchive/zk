package zk

// pingXID represents the XID which is used in ping/keepalive packet headers.
const pingXID = -2

// Below constants represent codes used by Zookeeper to differentiate requests.
// https://zookeeper.apache.org/doc/r3.4.8/api/constant-values.html#org.apache.zookeeper.ZooDefs.OpCode.getData
const (
	OpGetData     = 4
	OpGetChildren = 8
	opPing        = 11
)
