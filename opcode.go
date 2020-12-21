package udp

type OpCode int

const (
	OP_RECEIVE OpCode = 1
	OP_ERROR   OpCode = 2
)
