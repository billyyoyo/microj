package services

import (
	"context"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"github.com/billyyoyo/microj/logger"
	"google.golang.org/grpc"
	"time"
)

type TestService struct{}

func NewTestService() *TestService {
	s := &TestService{}
	return s
}

func (s *TestService) RegRpc(r *grpc.Server) {
	proto.RegisterTestServer(r, NewTestService())
}

func (s *TestService) Exec(ctx context.Context, req *proto.InParam) (resp *proto.OutParam, err error) {
	logger.Info("Exec recv: ", req.Value)
	t := time.Now().Format(time.DateTime)
	resp = &proto.OutParam{
		Msg: t,
	}
	return
}
