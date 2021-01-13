package network

import (
	"cloud/conf"
	"cloud/log"
	"cloud/network/protobuf"
	"context"
	"fmt"
	"net"
)

type Server struct {
	// 监听IP地址
	ip string
	// 绑定的端口号
	port uint32
	// 数据解析器
	msgParser *MsgParser
	// 数据包装
	processor *protobuf.Processor
	// ctx
	ctx context.Context
	cancelFunc context.CancelFunc
}

func NewServer(processor *protobuf.Processor) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{
		ip: conf.ServerConfig.IP,
		port: conf.ServerConfig.Port,
		msgParser:NewParser(),
		processor: processor,
		ctx: ctx,
		cancelFunc: cancel,
	}
	return s
}

func (s *Server)GetServerIP() string {
	return s.ip
}

func (s *Server)GetServerPort() uint32 {
	return s.port
}

// 启动服务器
func (s *Server)Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.GetServerIP(), s.GetServerPort()))
	defer listen.Close()

	if err != nil {
		log.Error("server listen err=%v", err)
		return
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Error("client conn err=%v", err)
			continue
		}

		dealSess := NewSession(s.ctx, conn, 1024, s.msgParser, s.processor)

		go dealSess.Start()
	}
}

// 关闭服务器
func (s *Server)Stop() {
	s.cancelFunc()
}