package request

import (
	"cloud/log"
	"runtime"
)

type Request struct {
	f func(...interface{})
	args []interface{}
}

func NewRequest(f func(...interface{}), args ...interface{}) Request {
	return Request{
		f:    f,
		args: args,
	}
}

func (r *Request) Call() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			log.Error("call func err=%s", buf[:l])
		}
	}()
	r.f(r.args...)
}
