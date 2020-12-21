package udp

import (
	"golang.org/x/sys/unix"
)

type Request struct {
	Id      uint64
	Op      OpCode
	From    unix.Sockaddr
	Msg     []byte
	N       int
	ErrCode ErrorCode
	Err     error
}

func requestRecycleUpdate(ptr interface{}) {
	var req, ok = ptr.(*Request)
	if ok && req != nil {
		resetRequest(req)
	}
}

func resetRequest(req *Request) {
	req.Msg = nil
}

func (ep *EP) getRequest() *Request {
	var req, err = ep.requestPool.Get()
	if err == nil {
		return req.(*Request)
	}
	return nil
}

func (ep *EP) putRequest(req *Request) {
	ep.requestPool.PutWithId(req, req.Id)
}

func (ep *EP) getRequestItem() *Request {
	return ep.getRequest()
}

func (ep *EP) InvokeReceive(from unix.Sockaddr, msg *[]byte, n int) {
	ep.threadPool.Invoke(ep.getRequestItemForReceive(from, msg, n))
}

func (ep *EP) InvokeError(code ErrorCode, err error) {
	ep.threadPool.Invoke(ep.getRequestItemForError(code, err))
}

func (ep *EP) getRequestItemForReceive(from unix.Sockaddr, msg *[]byte, n int) *Request {
	var req = ep.getRequestItem()
	req.Op = OP_RECEIVE
	req.From = from
	req.Msg = *msg
	req.N = n
	return req
}

func (ep *EP) getRequestItemForError(errCode ErrorCode, err error) *Request {
	var req = ep.getRequestItem()
	req.Op = OP_ERROR
	req.ErrCode = errCode
	req.Err = err
	return req
}
