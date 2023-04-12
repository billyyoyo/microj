package main

import (
	"context"
	"github.com/billyyoyo/microj/client"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/server/api"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Routers() []api.Api {
	return []api.Api{
		{Method: "get", Path: "/user", Func: UserInfo},
		{Method: "get", Path: "/say", Func: Say},
	}
}

func UserInfo(ctx *gin.Context) {
	userId := ctx.Query("id")
	cli := client.NewApiClient("server-api")
	resp, err := cli.R().
		SetHeader("Authorization", "abc").
		SetQueryParam("id", userId).
		Get("/info")
	if err != nil {
		logger.Error("call remote error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	ctx.Writer.Write(resp.Body())
}

func Say(ctx *gin.Context) {
	word := ctx.Query("word")
	conn := client.NewRpcConn("server-proto")
	cli := proto.NewExampleClient(conn)
	resp, err := cli.Call(context.Background(), &proto.Request{Value: word})
	if err != nil {
		logger.Error("call say error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"code": "0", "msg": "", "data": resp})
}
