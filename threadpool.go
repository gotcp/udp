package udp

import (
	"github.com/wuyongjia/threadpool"
)

func (ep *EP) newThreadPool() *threadpool.Pool {
	var p = threadpool.NewWithFunc(ep.Threads, ep.QueueLength, func(payload interface{}) {
		var req, ok = payload.(*Request)
		if ok {
			switch req.Op {
			case OP_RECEIVE:
				ep.OnReceive(req.From, req.Msg[:req.N], req.N)
				ep.PutBufferPoolItem(&req.Msg)
			case OP_ERROR:
				ep.OnError(req.ErrCode, req.Err)
			}
			ep.putRequest(req)
		}
	})
	return p
}
