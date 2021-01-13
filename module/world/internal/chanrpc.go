package internal

import (
	"cloud/pb"
	"fmt"
	"reflect"
)

func init() {
	skeleton.RegisterChanRPC("hello", helloTest)
	skeleton.RegisterChanRPC(reflect.TypeOf(&pb.CsPlayerInfo{}), PlayerInfo)
}

func helloTest(args []interface{}) {
	fmt.Println("int hello Test =", args)
}

func PlayerInfo(args []interface{}) {
	fmt.Println("in player info =", args)
}