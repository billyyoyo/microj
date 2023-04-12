package controllers

import (
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/examples/apisrv/codes"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/server/api"
	"github.com/gin-gonic/gin"
)

type ExampleController struct {
	api.BaseController
}

func NewExampleController() *ExampleController {
	return &ExampleController{}
}

func (c *ExampleController) Apis() []api.Api {
	return []api.Api{
		{Method: "post", Path: "/login", Func: c.Login},
		{Method: "get", Path: "/info", Func: c.GetInfo},
		{Method: "post", Path: "/edit", Func: c.Edit},
	}
}

func (c *ExampleController) Login(ctx *gin.Context) {
	var user struct {
		Name string `json:"username"`
		Pwd  string `json:"password"`
	}
	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		logger.Error("login params error", err)
		c.RetFailed(ctx, codes.ERR_REQEUST_PARAMS, "login params bind error")
		return
	}
	c.RetSuccess(ctx, gin.H{"msg": "ok"})
}

func (c *ExampleController) GetInfo(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	userId := ctx.Query("id")
	c.RetSuccess(ctx, gin.H{
		"token": token,
		"id":    userId,
		"name":  "Billyyoyo",
		"sex":   "1",
		"age":   "2",
	})
}

func (c *ExampleController) Edit(ctx *gin.Context) {
	var user struct {
		Id   string `json:"id"`
		Name string `json:"username"`
	}
	err := ctx.ShouldBindJSON(&user)
	if err != nil {
		c.RetFailed(ctx, codes.ERR_REQEUST_PARAMS, "params bind error")
		return
	}
	//var a int
	//b := 1 / a
	//ctx.JSON(http.StatusOK, gin.H{"msg": "ok", "ret": b})
	//logger.Panic("test", errors.New("test on error"))
	panic(errs.NewInternal("moring error"))
}
