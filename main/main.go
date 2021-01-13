package main

import (
	"cloud/module"
	"cloud/network"
	"cloud/route"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
)

func main() {
	// 日志相关设置
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// 协议转发
	s := network.NewServer(route.BaseProcessor)
	route.InitRegister()

	// 启动server
	go s.Start()

	// 启动module
	//module.Register(world.Module)
	module.Init()

	// 这里捕获退出信号 执行需要的退出操作
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c

	fmt.Println("server stop with sig=", sig)
	module.Destroy()
	s.Stop()
}