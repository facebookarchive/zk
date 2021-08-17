package opcodes

const (
	// PingXID represents the XID which is used in ping/keepalive packet headers
	PingXID = -2

	// OpGetData is the opcode for GetData requests
	OpGetData = 4
	// OpGetChildren is the opcode for GetChildren requests
	OpGetChildren = 8
	// OpPing is the opcode for ping/keepalive requests
	OpPing = 11
)
