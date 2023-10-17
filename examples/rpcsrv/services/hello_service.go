package services

import (
	"context"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/server/rpc"
	"google.golang.org/grpc"
	"time"
)

type HelloService struct{}

func NewHelloService() *HelloService {
	s := &HelloService{}
	return s
}

func (s *HelloService) RegRpc(r *grpc.Server) {
	proto.RegisterHelloServer(r, s)
}

func (s *HelloService) Say(ctx context.Context, req *proto.InParam) (resp *proto.OutParam, err error) {
	token, err := rpc.GetMetadata(ctx, "Authorization")
	if err != nil {
		return
	}
	logger.Info(token, "say recv: ", req.Value)
	t := time.Now().Format(time.DateTime)
	resp = &proto.OutParam{
		Msg: t,
	}
	err = errs.NewRpcError(510021, "can not say any thing")
	return
}
