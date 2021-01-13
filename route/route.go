package route

import (
	"cloud/chanrpc"
	"cloud/log"
	"cloud/mongodb"
	"cloud/network"
	"cloud/network/protobuf"
	"cloud/pb"
	"cloud/role"
	"cloud/util"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	BaseProcessor = protobuf.NewProcessor()
)

func InitRegister() {
	// 全部消息在此注册
	registerList := []proto.Message {
		&pb.CsHeart{},
		&pb.ScHeart{},
		&pb.CsPlayerCreate{},
		&pb.ScPlayerCreate{},
		&pb.CsAccountLogin{},
		&pb.ScAccountLogin{},
		&pb.CsPlayerInfo{},
		&pb.ScPlayerInfo{},
	}

	// 无需登录验证的消息
	sessionHandleList := []proto.Message{
		&pb.CsHeart{},
		&pb.CsPlayerCreate{},
		&pb.CsAccountLogin{},
	}

	// 需要进行登录验证的消息
	sessionRouteList := []struct{
		m proto.Message
		r *chanrpc.Server
	} {
		{&pb.CsPlayerInfo{}, nil},
	}

	for _, r := range registerList {
		BaseProcessor.Register(r)
	}

	for _, m := range sessionHandleList {
		BaseProcessor.SetHandler(m, SessionFunc)
	}
	for _, m := range sessionRouteList {
		BaseProcessor.SetHandlerRouter(m.m, HandleFunc, m.r)
	}
}

// session处理的消息统一通过此函数进行处理
func SessionFunc(args ...interface{}) {
	switch args[0].(type) {
	case *pb.CsHeart:
		CsHeart(args...)
	case *pb.CsPlayerCreate:
		CsPlayerCreate(args...)
	case *pb.CsAccountLogin:
		CsAccountLogin(args...)
	default:
		log.Error("error msg,err=%v", args)
	}
}

// 其余的消息需要进行登录状态判定 统一判断登录状态后再进行处理
func HandleFunc(args ...interface{}) {
	m := args[0]
	a := args[1].(*network.Session)
	agent := a.Property("agent")
	if agent == nil {
		a.Close()
		return
	}
	routeDest, msgType := BaseProcessor.RouteDest(m)
	if routeDest != nil {
		routeDest.Go(msgType, m, a)
	} else {
		agent.(*role.Agent).ChanRPCServer.Go(msgType, m, a, agent)
	}
}

// 心跳函数
func CsHeart(args ...interface{}) {
	a := args[1].(*network.Session)
	unixTime := util.NowTime()
	a.SetProperty("heart", unixTime)
	a.WriteMsg(&pb.ScHeart{UnixTime: unixTime})
}

// 注册函数
func CsPlayerCreate(args...interface{}) {
	// 收到的msg消息
	m := args[0].(*pb.CsPlayerCreate)

	a := args[1].(*network.Session)

	r := mongodb.BaseMongo.FindOne("g_role", bson.M{"name":m.Name})
	err := r.Err()
	switch err {
	case nil:
	case mongo.ErrNoDocuments:
	default:
		a.WriteMsg(&pb.ScPlayerCreate{Result: 0})
		return
	}

	_, err = mongodb.BaseMongo.InsertOne("g_role", &Player{
		RoleID: 1,
		Name:   m.Name,
	})
	if err != nil {
		a.WriteMsg(&pb.ScPlayerCreate{Result: 2})
	} else {
		a.WriteMsg(&pb.ScPlayerCreate{Result: 1})
	}
}

// 登录函数
func CsAccountLogin(args...interface{}) {
	// 收到的msg消息
	m := args[0].(*pb.CsAccountLogin)

	a := args[1].(*network.Session)

	r := mongodb.BaseMongo.FindOne("g_role", bson.M{"_id":m.RoleID})
	err := r.Err()
	switch err {
	case nil:
	case mongo.ErrNoDocuments:
		a.WriteMsg(&pb.ScAccountLogin{Result: 4})
		return
	default:
		a.WriteMsg(&pb.ScAccountLogin{Result: 3})
		return
	}
	p := &Player{}
	err = r.Decode(p)
	if err != nil {
		a.WriteMsg(&pb.ScAccountLogin{Result: 2})
	} else {
		agent := role.NewAgent(m.RoleID, p.Name)
		a.SetProperty("agent", agent)
		go agent.Run(a.Ctx)
		a.WriteMsg(&pb.ScAccountLogin{Result: 1})
	}
}