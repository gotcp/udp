package udp

import (
	"context"
	"net"
	"time"

	"golang.org/x/sys/unix"
)

func (ep *EP) Write(to unix.Sockaddr, msg []byte) error {
	return unix.Sendto(ep.Fd, msg, 0, to)
}

func (ep *EP) WriteWithTimeout(to unix.Sockaddr, msg []byte, timeout time.Duration) error {
	var ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var done = make(chan int8, 1)
	var err error
	go func() {
		err = unix.Sendto(ep.Fd, msg, 0, to)
		done <- 1
	}()
	select {
	case <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func GetAddr(addr unix.Sockaddr) (string, int) {
	switch addr.(type) {
	case *unix.SockaddrInet4:
		var v4 = addr.(*unix.SockaddrInet4)
		return net.IP(v4.Addr[:]).String(), v4.Port
	case *unix.SockaddrInet6:
		var v6 = addr.(*unix.SockaddrInet6)
		return net.IP(v6.Addr[:]).String(), v6.Port
	}
	return "", -1
}

func GetAddrBytes(addr unix.Sockaddr) ([]byte, int) {
	switch addr.(type) {
	case *unix.SockaddrInet4:
		var v4 = addr.(*unix.SockaddrInet4)
		return v4.Addr[:], v4.Port
	case *unix.SockaddrInet6:
		var v6 = addr.(*unix.SockaddrInet6)
		return v6.Addr[:], v6.Port
	}
	return nil, -1
}
