# Golang high-performance asynchronous UDP (only supports Linux)

## Example
 

```go
package main

import (
	"fmt"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/gotcp/udp"
)

var u *udp.EP

// Asynchronous event
func OnReceive(from unix.Sockaddr, msg []byte, n int) {
	var err = u.WriteWithTimeout(from, msg, 3*time.Second)
	if err != nil {
		fmt.Printf("write error -> %v\n", err)
	}
}

// Asynchronous event
func OnError(code udp.ErrorCode, err error) {
	fmt.Printf("OnError -> %d, %v\n", code, err)
}

func main() {
	var err error

	var rLimit syscall.Rlimit
	if err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	// parameters: readBuffer, threads, queueLength
	u, err = udp.New(2048, 3500, 4096)
	if err != nil {
		panic(err)
	}
	defer u.Stop()

	u.OnReceive = OnReceive // must have
	u.OnError = OnError     // optional

	u.Start("0.0.0.0", 8003)
}
```