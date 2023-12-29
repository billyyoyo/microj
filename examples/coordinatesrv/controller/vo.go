package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func SUCCESS(ctx *gin.Context, d interface{}) {
	ctx.JSON(http.StatusOK, Resp{
		Code: 0,
		Data: d,
	})
}

func FAILED(ctx *gin.Context, code int, msg string) {
	ctx.JSON(http.StatusOK, Resp{
		Code: code,
		Msg:  msg,
	})
}
