package services

import (
	"context"
	"github.com/billyyoyo/microj/broker"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"github.com/billyyoyo/microj/logger"
	"google.golang.org/grpc"
	"time"
)

type ExampleService struct{}

func NewExampleService() *ExampleService {
	s := &ExampleService{}
	return s
}

func (s *ExampleService) RegRpc(r *grpc.Server) {
	proto.RegisterExampleServer(r, NewExampleService())
}

func (s *ExampleService) Call(ctx context.Context, req *proto.Request) (resp *proto.Response, err error) {
	logger.Info("call recv: ", req.Value)
	t := time.Now().Format(time.DateTime)
	resp = &proto.Response{
		Msg: t,
	}
	broker.Send("example-hello", broker.Message{Body: []byte(t)})
	return
}
