package network

import (
	"encoding/binary"
	"errors"
	"io"
)

type MsgParser struct {
	// 消息体长度长度
	lenMsgLen int
	minMsgLen    uint32
	maxMsgLen    uint32
	// 大小端
	littleEndian bool
}

func NewParser() *MsgParser {
	return  &MsgParser{
		lenMsgLen: 4,
		minMsgLen: 1,
		maxMsgLen: 4096,
		littleEndian: false,
	}
}

func (p *MsgParser) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

func (p *MsgParser) Read(conn io.Reader) ([]byte, error) {
	var b [4]byte
	bufMsgLen := b[:p.lenMsgLen]

	// 读取消息体长度字段
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err
	}

	// 解析长度字段
	var msgLen uint32
	switch p.lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		if p.littleEndian {
			msgLen = uint32(binary.LittleEndian.Uint16(bufMsgLen))
		} else {
			msgLen =uint32(binary.BigEndian.Uint16(bufMsgLen))
		}
	case 4:
		if p.littleEndian {
			msgLen = uint32(binary.LittleEndian.Uint32(bufMsgLen))
		} else {
			msgLen =uint32(binary.BigEndian.Uint32(bufMsgLen))
		}
	}

	// 判断长度是否合法
	if msgLen > p.maxMsgLen {
		return nil, errors.New("msg too long")
	}
	if msgLen < p.minMsgLen {
		return nil, errors.New("msg too short")
	}

	// 读取数据
	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}

	return msgData, nil
}

func (p *MsgParser) Write(conn *Session, args ...[]byte) error {
	msg, err := p.Analysis(args...)
	if err != nil {
		return err
	}
	conn.Write(msg)

	return nil
}

func (p *MsgParser) Analysis(args ...[]byte) ([]byte, error){
	// 获取消息长度
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// 判断长度是否合法
	if msgLen > p.maxMsgLen {
		return nil, errors.New("msg too long")
	}
	if msgLen < p.minMsgLen {
		return nil, errors.New("msg too short")
	}

	msg := make([]byte, uint32(p.lenMsgLen)+msgLen)

	// 写入长度信息
	switch p.lenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		if p.littleEndian {
			binary.LittleEndian.PutUint16(msg, uint16(msgLen))
		} else {
			binary.BigEndian.PutUint16(msg, uint16(msgLen))
		}
	case 4:
		if p.littleEndian {
			binary.LittleEndian.PutUint32(msg, msgLen)
		} else {
			binary.BigEndian.PutUint32(msg, msgLen)
		}
	}

	// 写入数据
	l := p.lenMsgLen
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	return msg, nil
}



