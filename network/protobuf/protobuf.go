package protobuf

import (
	"cloud/chanrpc"
	"cloud/log"
	"cloud/network/request"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"math"
	"reflect"
)

type Processor struct {
	littleEndian bool
	routerMap map[uint16]*msgInfo
	msgID	map[reflect.Type]uint16
}

type msgInfo struct {
	msgType reflect.Type
	msgHandler MsgHandler
	msgRouter *chanrpc.Server
}

type MsgHandler func(...interface{})

func NewProcessor() *Processor {
	p := new(Processor)
	p.littleEndian = false
	p.msgID = make(map[reflect.Type]uint16)
	p.routerMap = make(map[uint16]*msgInfo)
	return p
}

func (p *Processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

func (p *Processor) Register(msg proto.Message) {
	msgType := reflect.TypeOf(msg)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		return
	}

	if _, ok := p.msgID[msgType]; ok {
		return
	}

	if len(p.routerMap) >= math.MaxUint32 {
		return
	}

	i := new(msgInfo)
	i.msgType = msgType
	p.routerMap[uint16(len(p.msgID))] = i
	p.msgID[msgType] = uint16(len(p.routerMap)-1)
}

// 该类消息传给对应module直接处理
func (p *Processor) SetRouter(msg proto.Message, msgRouter *chanrpc.Server) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Panic("message %s not registered", msgType)
	}

	p.routerMap[id].msgRouter = msgRouter
}

// 该类消息传给对应连接直接处理
func (p *Processor) SetHandler(msg proto.Message, handler MsgHandler) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		return
	}
	p.routerMap[id].msgHandler = handler
}

// 同时设置handle and route
func (p *Processor) SetHandlerRouter(msg proto.Message, handler MsgHandler, msgRouter *chanrpc.Server) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		log.Panic("message %s not registered", msgType)
	}
	p.routerMap[id].msgHandler = handler
	p.routerMap[id].msgRouter = msgRouter
}

// 传递消息给同一的消息队列按序处理
func (p *Processor) Route(msg interface{}, userData interface{}, Chan chan request.Request, directly bool) error {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		return fmt.Errorf("message %s not registered", msgType)
	}

	i := p.routerMap[id]
	if i.msgHandler != nil {
		Chan <- request.NewRequest(i.msgHandler, msg, userData)
	}

	if directly && i.msgRouter != nil {
		i.msgRouter.Go(msgType, msg, userData)
	}
	return nil
}

// 获取消息最终传递位置
func (p *Processor) RouteDest(msg interface{}) (*chanrpc.Server, reflect.Type) {
	msgType := reflect.TypeOf(msg)
	id, ok := p.msgID[msgType]
	if !ok {
		return nil, nil
	}
	i := p.routerMap[id]
	if i.msgRouter != nil {
		return i.msgRouter, msgType
	}
	return nil, msgType
}

func (p *Processor) Marshal(msg interface{}) ([][]byte, error) {
	msgType := reflect.TypeOf(msg)

	// id
	_id, ok := p.msgID[msgType]
	if !ok {
		err := fmt.Errorf("message %s not registered", msgType)
		return nil, err
	}

	id := make([]byte, 2)
	if p.littleEndian {
		binary.LittleEndian.PutUint16(id, _id)
	} else {
		binary.BigEndian.PutUint16(id, _id)
	}

	data, err := proto.Marshal(msg.(proto.Message))
	return [][]byte{id, data}, err
}

func (p *Processor) Unmarshal(data []byte) (interface{}, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	var id uint16
	if p.littleEndian {
		id = binary.LittleEndian.Uint16(data)
	} else {
		id = binary.BigEndian.Uint16(data)
	}

	if id >= uint16(len(p.routerMap)) {
		return nil, fmt.Errorf("message id %v not registered", id)
	}
	msg := reflect.New(p.routerMap[id].msgType.Elem()).Interface()
	return msg, proto.UnmarshalMerge(data[2:], msg.(proto.Message))
}