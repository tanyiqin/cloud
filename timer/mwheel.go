package timer

import (
	"cloud/log"
	"cloud/util"
	"context"
	"sync"
	"time"
)

type Timer struct {
	// 执行时间
	unixTime int64
	// 执行函数
	callBackFunc func(...interface{})
	// 函数执行的参数
	args []interface{}
	// tid
	tid uint32
}

type TimerWheel struct {
	// tid
	tid uint32
	mutex sync.Mutex
	waitGroup util.WrapWait
	// 通过对时间hash 找到对应时间所要执行的所有Timer map
	timers map[int64]map[uint32]*Timer
	// 上次执行到的时间
	preTime int64
	// tid->时间的映射
	tidMapUnixTime map[uint32]int64
	// ctx
	ctx context.Context
}

func (t *Timer) doFunc() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("timer doFunc error=", err)
		}
	}()
	t.callBackFunc(t.args...)
}

func NewMWheel(ctx context.Context) *TimerWheel{
	return &TimerWheel{
		timers : make(map[int64]map[uint32]*Timer, 100),
		tid: 1,
		waitGroup: util.WrapWait{},
		tidMapUnixTime : make(map[uint32]int64, 100),
		ctx: ctx,
	}
}

func (h *TimerWheel)AddTimer(unixTime int64, callBackFunc func(...interface{}), args... interface{}) uint32{
	if h.preTime >= unixTime {
		return 0
	}
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if _, ok := h.timers[unixTime]; !ok {
		h.timers[unixTime] = make(map[uint32]*Timer, 10)
	}
	h.timers[unixTime][h.tid] = &Timer{
		unixTime:     unixTime,
		callBackFunc: callBackFunc,
		args:         args,
		tid:		  h.tid,
	}
	h.tidMapUnixTime[h.tid] = unixTime
	h.tid++
	return h.tid-1
}

func (h *TimerWheel)AddTicker(period int64, callBackFunc func(...interface{}), args... interface{}) uint32{
	nextTime := period + util.NowTime()
	callBackFunc2 := func(args ...interface{}) {
		callBackFunc(args...)
		h.AddTicker(period, callBackFunc, args...)
	}
	return h.AddTimer(nextTime, callBackFunc2, args...)
}

func (h *TimerWheel) Run() {
	h.preTime = util.NowTime()
	go tick(h)
}

func (h *TimerWheel) StopTimer(tid uint32){
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if unixTime, ok := h.tidMapUnixTime[tid]; ok {
		delete(h.tidMapUnixTime, tid)
		if _, ok := h.timers[unixTime][tid]; ok {
			delete(h.timers[unixTime], tid)
		}
	}
}

func (h *TimerWheel) Stop() {
	h.waitGroup.Wait()
}

func tick(h *TimerWheel) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	tChan := time.After(1*time.Second)
	CurrTime := util.NowTime()
	for i := h.preTime + 1; i <= CurrTime; i++ {
		if m, ok := h.timers[i]; ok {
			for _, t := range m {
				go h.waitGroup.Wrap(t.doFunc)
				delete(h.tidMapUnixTime, t.tid)
			}
		}
		delete(h.timers, i)
	}
	h.preTime = CurrTime
	go func() {
		select {
		case <-tChan:
			tick(h)
		case <- h.ctx.Done():
			return
		}
	}()
}