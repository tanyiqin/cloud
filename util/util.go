package util

import (
	"sync"
	"time"
)

// 当前时间/s
func NowTime() int64{
	t := time.Now()
	return t.Unix()
}

// 用来等待所有go协程返回
type WrapWait struct {
	sync.WaitGroup
}

func (w *WrapWait)Wrap(f func()) {
	w.Add(1)
	go func() {
		f()
		w.Done()
	}()
}
