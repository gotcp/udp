package udp

import (
	"golang.org/x/sys/unix"
)

type OnReceiveEvent func(from unix.Sockaddr, msg []byte, n int)
type OnErrorEvent func(code ErrorCode, err error)
