package module

import (
	"cloud/chanrpc"
	"cloud/log"
	"context"
)

type Skeleton struct {
	ChanRPCServer      *chanrpc.Server
}

func NewSkeleton() *Skeleton {
	skeleton := &Skeleton{
		ChanRPCServer: chanrpc.NewServer(1024),
	}

	return skeleton
}

func (s *Skeleton) Init() {
}

func (s *Skeleton) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			s.ChanRPCServer.Close()
			return
		case ci := <-s.ChanRPCServer.ChanCall:
			err := s.ChanRPCServer.Exec(ci)
			if err != nil {
				log.Error("%v", err)
			}
		}
	}
}

func (s *Skeleton) RegisterChanRPC(id interface{}, f interface{}) {
	if s.ChanRPCServer == nil {
		panic("invalid ChanRPCServer")
	}

	s.ChanRPCServer.Register(id, f)
}
