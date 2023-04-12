package main

import (
	"context"
	"fmt"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"google.golang.org/grpc"
	"testing"
)

func TestRpc(t *testing.T) {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		panic("grpc start up error: " + err.Error())
		return
	}
	cli := proto.NewExampleClient(conn)
	req := &proto.Request{
		Value: "hello world",
	}
	resp, err := cli.Call(context.Background(), req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(resp.String())
}
