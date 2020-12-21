package udp

import (
	"net"
	"time"

	"golang.org/x/sys/unix"

	"github.com/wuyongjia/pool"
	"github.com/wuyongjia/threadpool"
)

type EP struct {
	Host        string
	Port        int
	Fd          int
	Threads     int
	QueueLength int
	ReadBuffer  int
	ReuseAddr   int
	ReusePort   int
	delay       time.Duration
	bufferPool  *pool.Pool       // []byte pool, return *[]byte
	threadPool  *threadpool.Pool // thread pool sequence
	requestPool *pool.Pool       // *Request pool, return *Request
	OnReceive   OnReceiveEvent
	OnError     OnErrorEvent
}

const (
	DEFAULT_EPOLL_EVENTS        = 4096
	DEFAULT_EPOLL_READ_TIMEOUT  = 5
	DEFAULT_EPOLL_WRITE_TIMEOUT = 5
	DEFAULT_POOL_MULTIPLE       = 5
	DEFAULT_DELAY               = 200
)

func New(readBuffer int, threads int, queueLength int) (*EP, error) {
	var ep = &EP{
		Fd:          -9,
		ReadBuffer:  readBuffer,
		Threads:     threads,
		QueueLength: queueLength,
		ReuseAddr:   1,
		ReusePort:   1,
		delay:       DEFAULT_DELAY * time.Millisecond,
		OnReceive:   nil,
		OnError:     nil,
	}

	ep.bufferPool = ep.newBufferPool(readBuffer, threads*DEFAULT_POOL_MULTIPLE)
	ep.requestPool = ep.newRequestPool(threads * DEFAULT_POOL_MULTIPLE)

	ep.bufferPool.RecycleUpdateFunc = bufferRecycleUpdate
	ep.requestPool.RecycleUpdateFunc = requestRecycleUpdate

	ep.bufferPool.EnableRecycle()
	ep.requestPool.EnableRecycle()

	ep.threadPool = ep.newThreadPool()

	return ep, nil
}

func (ep *EP) newBufferPool(readBuffer int, length int) *pool.Pool {
	return pool.New(length, func() interface{} {
		var b = make([]byte, readBuffer)
		return &b
	})
}

func (ep *EP) newRequestPool(length int) *pool.Pool {
	return pool.NewWithId(length, func(id uint64) interface{} {
		return &Request{Id: id}
	})
}

func (ep *EP) SetDelay(t time.Duration) {
	ep.delay = t
}

func (ep *EP) SetReuseAddr(n int) {
	ep.ReuseAddr = n
}

func (ep *EP) SetReusePort(n int) {
	ep.ReusePort = n
}

func (ep *EP) Start(host string, port int) {
	var err error
	if err = ep.initEpoll(host, port); err != nil {
		panic(err)
	}
	ep.listen()
}

func (ep *EP) Stop() {
	unix.Close(ep.Fd)
	ep.threadPool.Close()
}

func (ep *EP) initEpoll(host string, port int) error {
	var err error

	if ep.Fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_UDP); err != nil {
		return err
	}

	if err = unix.SetsockoptInt(ep.Fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, ep.ReuseAddr); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	if err = unix.SetsockoptInt(ep.Fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, ep.ReusePort); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	var addr = unix.SockaddrInet4{Port: port}
	copy(addr.Addr[:], net.ParseIP(host).To4())

	if err = unix.Bind(ep.Fd, &addr); err != nil {
		unix.Close(ep.Fd)
		return err
	}

	return nil
}

func (ep *EP) listen() {
	var err error
	var n int
	var msg *[]byte
	var from unix.Sockaddr
	for {
		msg, err = ep.GetBufferPoolItem()
		if err != nil {
			if ep.OnError != nil {
				ep.InvokeError(ERROR_BUFFER_POOL, err)
			}
			time.Sleep(ep.delay)
			continue
		}
		n, from, err = unix.Recvfrom(ep.Fd, *msg, 0)
		if err != nil {
			ep.PutBufferPoolItem(msg)
			if ep.OnError != nil {
				ep.InvokeError(ERROR_RECEIVE, err)
			}
			break
		}
		ep.InvokeReceive(from, msg, n)
	}
}
