package zk

const (
	// PermAll is an ACL permission which permits all operation on a node.
	PermAll = 0x1f

	defaultProtocolVersion = 0

	pingXID = -2

	opCreate      = 1
	opGetData     = 4
	opGetChildren = 8

	opPing = 11
)
