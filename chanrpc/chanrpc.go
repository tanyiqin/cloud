package chanrpc

import (
	"errors"
	"fmt"
	"runtime"
)

// one server per goroutine (goroutine not safe)
// one client per goroutine (goroutine not safe)
type Server struct {
	*FuncManager
	ChanCall  chan *CallInfo
}

type FuncManager struct {
	// id -> function
	//
	// function:
	// func(args []interface{})
	// func(args []interface{}) interface{}
	// func(args []interface{}) []interface{}
	functions map[interface{}]interface{}
}

type CallInfo struct {
	f       interface{}
	args    []interface{}
	chanRet chan *RetInfo
}

type RetInfo struct {
	// nil
	// interface{}
	// []interface{}
	ret interface{}
	err error
}

type Client struct {
	s               *Server
	chanSyncRet     chan *RetInfo
}

func NewFuncManager() *FuncManager{
	s := new(FuncManager)
	s.functions = make(map[interface{}]interface{})
	return s
}

func NewServer(l int) *Server {
	s := new(Server)
	s.ChanCall = make(chan *CallInfo, l)
	return s
}

// you must call the function before calling Open and Go
func (s *FuncManager) Register(id interface{}, f interface{}) {
	switch f.(type) {
	case func([]interface{}):
	case func([]interface{}) interface{}:
	case func([]interface{}) []interface{}:
	default:
		panic(fmt.Sprintf("function id %v: definition of function is invalid", id))
	}

	if _, ok := s.functions[id]; ok {
		panic(fmt.Sprintf("function id %v: already registered", id))
	}

	s.functions[id] = f
}

func (s *Server) ret(ci *CallInfo, ri *RetInfo) (err error) {
	if ci.chanRet == nil {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	ci.chanRet <- ri
	return
}

func (s *Server) Exec(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 1024)
			l := runtime.Stack(buf, false)
			err = fmt.Errorf("%v: %s", r, buf[:l])
			s.ret(ci, &RetInfo{err: fmt.Errorf("%v", r)})
		}
	}()

	// execute
	switch ci.f.(type) {
	case func([]interface{}):
		ci.f.(func([]interface{}))(ci.args)
		return s.ret(ci, &RetInfo{})
	case func([]interface{}) interface{}:
		ret := ci.f.(func([]interface{}) interface{})(ci.args)
		return s.ret(ci, &RetInfo{ret: ret})
	case func([]interface{}) []interface{}:
		ret := ci.f.(func([]interface{}) []interface{})(ci.args)
		return s.ret(ci, &RetInfo{ret: ret})
	}

	panic("bug")
}

// goroutine safe
func (s *Server) Go(id interface{}, args ...interface{}) {
	f := s.functions[id]
	if f == nil {
		return
	}

	defer func() {
		recover()
	}()

	s.ChanCall <- &CallInfo{
		f:    f,
		args: args,
	}
}

func (s *Server) Close() {
	close(s.ChanCall)

	for ci := range s.ChanCall {
		s.ret(ci, &RetInfo{
			err: errors.New("chanrpc server closed"),
		})
	}
}

// goroutine safe
func (s *Server) Open() *Client {
	c := new(Client)
	c.s = s
	c.chanSyncRet = make(chan *RetInfo, 1)
	return c
}
func (s *Server) Open1(chanSyncRet chan *RetInfo) *Client {
	c := new(Client)
	c.s = s
	c.chanSyncRet = chanSyncRet
	return c
}

func (c *Client) call(ci *CallInfo) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	c.s.ChanCall <- ci
	return nil
}

func (c *Client) f(id interface{}, n int) (f interface{}, err error) {
	f = c.s.functions[id]
	if f == nil {
		err = fmt.Errorf("function id %v: function not registered", id)
		return
	}

	var ok bool
	switch n {
	case 0:
		_, ok = f.(func([]interface{}))
	case 1:
		_, ok = f.(func([]interface{}) interface{})
	case 2:
		_, ok = f.(func([]interface{}) []interface{})
	default:
		panic("bug")
	}

	if !ok {
		err = fmt.Errorf("function id %v: return type mismatch", id)
	}
	return
}

func (c *Client) Call0(id interface{}, args ...interface{}) error {
	f, err := c.f(id, 0)
	if err != nil {
		return err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.chanSyncRet,
	})
	if err != nil {
		return err
	}

	ri := <-c.chanSyncRet
	return ri.err
}

func (c *Client) Call1(id interface{}, args ...interface{}) (interface{}, error) {
	f, err := c.f(id, 1)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.chanSyncRet,
	})
	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet
	return ri.ret, ri.err
}

func (c *Client) CallN(id interface{}, args ...interface{}) ([]interface{}, error) {
	f, err := c.f(id, 2)
	if err != nil {
		return nil, err
	}

	err = c.call(&CallInfo{
		f:       f,
		args:    args,
		chanRet: c.chanSyncRet,
	})
	if err != nil {
		return nil, err
	}

	ri := <-c.chanSyncRet
	return ri.ret.([]interface{}), ri.err
}

func (c *Client) Close() {

}