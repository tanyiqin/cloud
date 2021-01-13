package network

import (
	"cloud/log"
	"cloud/network/protobuf"
	"cloud/network/request"
	"cloud/util"
	"context"
	"io"
	"net"
	"sync"
	"time"
)

type Session struct {
	// 连接套接字
	Conn net.Conn
	// 数据解析器
	msgParser *MsgParser
	// 发送消息的缓存
	writeChan chan []byte
	// 连接是否关闭
	closeFlag bool
	sync.Mutex
	// 数据包装
	processor *protobuf.Processor
	// ctx
	Ctx context.Context
	cancelFunc context.CancelFunc
	// 消息队列 所有的消息都从这里取出来处理
	requestChan chan request.Request
	// 各种状态信息
	property map[interface{}]interface{}
}

func NewSession(ctx context.Context, Conn net.Conn, writeChanLen uint32, msgParser *MsgParser, processor *protobuf.Processor) *Session {
	ctx1, cancel := context.WithCancel(ctx)
	return &Session{
		Conn: Conn,
		msgParser: msgParser,
		writeChan: make(chan []byte, writeChanLen),
		processor: processor,
		Ctx: ctx1,
		cancelFunc: cancel,
		requestChan: make(chan request.Request, 1024),
		property: make(map[interface{}]interface{}),
	}
}

func (c *Session) Start() {
	go c.StartReader()
	go c.StartWriter()

	ticker := time.NewTicker(10*time.Second)

	for {
		select {
		case <- c.Ctx.Done():
			return
		case r, ok := <- c.requestChan:
			if ok {
				r.Call()
			}
		case <- ticker.C:
			if heart := c.Property("heart"); heart != nil {
				if heart.(int64) + int64(30*time.Second) < util.NowTime() {
					c.Close()
				}
			} else {
				c.SetProperty("heart", util.NowTime())
			}
		}
	}
}

// 用于从socket读取数据并转交
func (c *Session) StartReader() {
	defer func() {
		close(c.requestChan)
		c.Close()
	}()

	for {
		data, err := c.msgParser.Read(c)
		if err != nil && err != io.EOF{
			log.Error("parser data err=%T %+v", err, err)
			return
		}
		msg, err := c.processor.Unmarshal(data)
		if err != nil {
			log.Error("marshal data err=%v", err)
			return
		}
		err = c.processor.Route(msg, c, c.requestChan, false)
		if err != nil {
			log.Error("route msg err=%v", err)
			break
		}
	}
}

// 用于将消息发给client
func (c *Session) StartWriter() {
	defer c.Close()
	for {
		select {
		case data := <-c.writeChan:
			c.Conn.Write(data)
		case <- c.Ctx.Done():
			return
		}
	}
}

func (c *Session) Read(b []byte) (int, error) {
	return c.Conn.Read(b)
}

// 发送消息的接口
func (c *Session) WriteMsg(msg interface{}) {
	data, err := c.processor.Marshal(msg)
	if err != nil {
		log.Error("marshal data err=%v", err)
		return
	}
	if err := c.msgParser.Write(c, data...); err != nil {
		log.Error("write data err=%v", err)
		return
	}
}

// 消息将会交付给写协程单独处理
func (c *Session) Write(b []byte) {
	c.Lock()
	defer c.Unlock()
	if c.closeFlag {
		return
	}
	if len(c.writeChan) == cap(c.writeChan) {
		c.Close()
		return
	}
	c.writeChan <- b
}

func (c *Session) SetProperty(key interface{}, val interface{}) {
	c.property[key] = val
}

func (c *Session) Property(key interface{}) interface{} {
	if val, ok := c.property[key]; ok {
		return val
	}
	return nil
}

func (c *Session) Close() {
	if c.closeFlag {
		return
	}
	c.closeFlag = true
	//c.wm.Stop()
	//c.doPersist()
	c.cancelFunc()
	c.Conn.Close()
}