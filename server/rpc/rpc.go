package rpc

import (
	"context"
	"fmt"
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"net"
	"strings"
)

type RpcReg func(r *grpc.Server)

type RpcServer struct {
	rpcReg []RpcReg
	s      *grpc.Server
}

func (s *RpcServer) RegService(reg RpcReg) *RpcServer {
	s.rpcReg = append(s.rpcReg, reg)
	return s
}

func (s *RpcServer) Init() {

}

func (s *RpcServer) Run() {
	logger.Info("init proto server")
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", app.Addr(), app.Port()))
	if err != nil {
		logger.Fatal("grpc server listen error:", err)
	}
	rpcServer := grpc.NewServer()
	for _, s := range s.rpcReg {
		s(rpcServer)
	}
	if err = rpcServer.Serve(listener); err != nil {
		logger.Fatal("grpc server startup error:", err)
	}
}

func (s *RpcServer) Stop() {
	s.s.GracefulStop()
}

func GetMetadata(ctx context.Context, key string) (value string, err error) {
	vs := metadata.ValueFromIncomingContext(ctx, strings.ToLower(key))
	if vs == nil || len(vs) == 0 {
		err = errs.NewRpcError(errs.ERRCODE_NO_TOKEN, fmt.Sprintf("not found incoming param %s", key))
		return
	}
	value = vs[0]
	return
}
