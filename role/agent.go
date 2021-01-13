package role

import (
	"cloud/chanrpc"
	"cloud/module"
	"cloud/network"
	"cloud/pb"
	"github.com/golang/protobuf/proto"
	"reflect"
)

var (
	AgentFuncManager = chanrpc.NewFuncManager()
)

func init() {
	// 玩家进程注册
	r1 := [] struct{
		p proto.Message
		f interface{}
	} {
		{&pb.CsPlayerInfo{}, CsPlayerInfo},
	}
	for _, r := range r1 {
		AgentFuncManager.Register(reflect.TypeOf(r.p), r.f)
	}
}

type Agent struct {
	*module.Skeleton
	roleID uint32
	roleName string
	property map[interface{}]interface{}
}

func NewAgent(roleID uint32, roleName string) *Agent {
	a := &Agent{
		roleID:   roleID,
		roleName: roleName,
	}
	a.Skeleton = module.NewSkeleton()
	a.ChanRPCServer.FuncManager = AgentFuncManager
	return a
}

func CsPlayerInfo(args []interface{}) {
	_ = args[0].(*pb.CsPlayerInfo)
	a := args[1].(*network.Session)
	agent := a.Property("agent").(*Agent)
	a.WriteMsg(&pb.ScPlayerInfo{
		Result:               1,
		Name:                 agent.roleName,
		RoleID:               agent.roleID,
	})
}