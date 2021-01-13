package main

import (
	"cloud/route"
	"cloud/network"
	"cloud/pb"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("conn err,", err)
		return
	}
	route.InitRegister()
	msgList := []proto.Message {
		//&pb.CsPlayerCreate{Name: "wewe"},
		&pb.CsAccountLogin{RoleID: 1},
		&pb.CsPlayerInfo{},
	}
	for _, msg := range msgList {
		Parser := network.NewParser()
		data, err := route.BaseProcessor.Marshal(msg)
		if err != nil {
			fmt.Println("data err, err=", err)
			return
		}
		msg1, err := Parser.Analysis(data...)
		if err != nil {
			fmt.Println("msg1 err=",err)
			return
		}
		_, err = conn.Write(msg1)
		if err != nil {
			fmt.Println("conn write err,err=",err)
			return
		}

		data1, err := Parser.Read(conn)
		if err != nil {
			fmt.Println("data1 err,err=",err)
			return
		}
		msg2, err := route.BaseProcessor.Unmarshal(data1)
		if err != nil {
			fmt.Println("msg2 err,err=",err)
			return
		}
		fmt.Println(msg2)
	}
	//time.Sleep(50*time.Second)
}
