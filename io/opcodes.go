package io

const (
	// PingXID represents the XID which is used in ping/keepalive packet headers.
	PingXID = -2

	// OpGetData and other constants with the same prefix represent codes used by Zookeeper to differentiate requests.
	// reference: https://zookeeper.apache.org/doc/r3.4.8/api/constant-values.html#org.apache.zookeeper.ZooDefs.OpCode.getData
	OpGetData     = 4
	OpGetChildren = 8
	OpPing        = 11
)
