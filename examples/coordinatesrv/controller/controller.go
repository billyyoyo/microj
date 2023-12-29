package controller

import (
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/examples/coordinatesrv/coor"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/server/api"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io"
	"time"
)

var (
	PATHS = []string{"/data/set", "/data/del", "/admin/join", "/admin/remove"}
)

type CoordController struct {
	handler *coor.CoordinatorHandler
}

func New(r *coor.CoordinatorHandler) *CoordController {
	return &CoordController{
		handler: r,
	}
}

func (c *CoordController) ProxyFilter(ctx *gin.Context) {
	var forward bool
	if c.handler.Raft.State() != raft.Leader {
		for _, p := range PATHS {
			if p == ctx.Request.URL.Path {
				forward = true
				break
			}
		}
	}
	if forward {
		defer ctx.Abort()
		leaderAddr, err := c.handler.GetLeaderApiAddress()
		if err != nil {
			logger.Err(errors.Wrap(err, "no leader address can usage"))
			FAILED(ctx, 500, "no leader address can usage")
			return
		}
		cli := resty.New()
		cli.SetBaseURL(fmt.Sprintf("http://%s", leaderAddr))
		req := cli.R()
		req.Method = ctx.Request.Method
		req.URL = ctx.Request.URL.Path
		req.SetHeaderMultiValues(ctx.Request.Header)
		req.SetQueryString(ctx.Request.URL.RawQuery)
		if ctx.Request.Method == "POST" {
			defer func() {
				logger.Info("Close req body")
				ctx.Request.Body.Close()
			}()
			bs, _ := io.ReadAll(ctx.Request.Body)
			logger.Info(string(bs))
			req.SetBody(bs)
		}
		resp, err := req.Send()
		if err != nil {
			err = errors.Wrap(err, "request forward failed")
			FAILED(ctx, 500, "request forward failed")
			return
		}
		ctx.Data(resp.StatusCode(), "application/json", resp.Body())
		logger.Info("call leader api finished")
	} else {
		logger.Info("before exec")
		ctx.Next()
		logger.Info("after exec")
	}
}

func (c *CoordController) Apis() []api.Api {
	return []api.Api{
		{
			Method: "get",
			Path:   "/data/get",
			Func:   c.Get,
		},
		{
			Method: "post",
			Path:   "/data/set",
			Func:   c.Set,
		},
		{
			Method: "post",
			Path:   "/data/del",
			Func:   c.Del,
		},
		{
			Method: "post",
			Path:   "/admin/join",
			Func:   c.PeerJoin,
		},
		{
			Method: "get",
			Path:   "/admin/leader",
			Func:   c.PeerLeader,
		},
		{
			Method: "post",
			Path:   "/admin/remove",
			Func:   c.PeerRemove,
		},
		{
			Method: "get",
			Path:   "/admin/stats",
			Func:   c.PeerStats,
		},
	}
}

func (c *CoordController) Get(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		FAILED(ctx, 400, "need name parameter")
		return
	}
	value := c.handler.Get(name)
	SUCCESS(ctx, value)
}

func (c *CoordController) Set(ctx *gin.Context) {
	form := make(map[string]interface{})
	err := ctx.ShouldBindJSON(&form)
	if err != nil {
		logger.Err(err)
		FAILED(ctx, 400, fmt.Sprintf("parse param error: %s", err.Error()))
		return
	}
	cmd := coor.Command{
		Action:  "set",
		Payload: form,
	}
	bs, _ := json.Marshal(&cmd)
	future := c.handler.Raft.Apply(bs, time.Second)
	if future.Error() != nil {
		FAILED(ctx, 500, fmt.Sprintf("raft apply log failed: %s", future.Error().Error()))
		return
	}
	if err, ok := future.Response().(error); ok {
		FAILED(ctx, 500, fmt.Sprintf("raft apply log resp failed: %s", err.Error()))
		return
	}
	SUCCESS(ctx, nil)
}

func (c *CoordController) Del(ctx *gin.Context) {
	name := ctx.Query("name")
	if name == "" {
		FAILED(ctx, 400, "need name parameter")
		return
	}
	cmd := coor.Command{
		Action: "del",
		Payload: map[string]interface{}{
			"key": name,
		},
	}
	bs, _ := json.Marshal(&cmd)
	future := c.handler.Raft.Apply(bs, time.Second)
	if future.Error() != nil {
		FAILED(ctx, 500, errors.Wrap(future.Error(), "raft apply log failed").Error())
		return
	}
	SUCCESS(ctx, nil)
}

func (c *CoordController) PeerLeader(ctx *gin.Context) {
	addr, id := c.handler.Raft.LeaderWithID()
	SUCCESS(ctx, map[string]string{
		"address": string(addr),
		"id":      string(id),
	})
}

type JoinPeerReq struct {
	Id   string `json:"id"`
	Addr string `json:"addr"`
}

func (c *CoordController) PeerJoin(ctx *gin.Context) {
	req := JoinPeerReq{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		FAILED(ctx, 400, errors.Wrap(err, "parse join request failed").Error())
		return
	}
	if req.Id == "" || req.Addr == "" {
		FAILED(ctx, 400, "id and address must not be empty")
		return
	}
	confFuture := c.handler.Raft.GetConfiguration()
	if err = confFuture.Error(); err != nil {
		FAILED(ctx, 500, errors.Wrap(err, "can not get raft configure").Error())
		return
	}
	future := c.handler.Raft.AddVoter(raft.ServerID(req.Id), raft.ServerAddress(req.Addr), 0, 0)
	if err = future.Error(); err != nil {
		FAILED(ctx, 500, errors.Wrap(err, "can not join node").Error())
		return
	}
	SUCCESS(ctx, nil)
}

type RemovePeerReq struct {
	Id string `json:"id"`
}

func (c *CoordController) PeerRemove(ctx *gin.Context) {
	req := RemovePeerReq{}
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		FAILED(ctx, 400, errors.Wrap(err, "parse join request failed").Error())
		return
	}
	if req.Id == "" {
		FAILED(ctx, 400, "id and address must not be empty")
		return
	}
	confFuture := c.handler.Raft.GetConfiguration()
	if err = confFuture.Error(); err != nil {
		FAILED(ctx, 500, errors.Wrap(err, "can not get raft configure").Error())
		return
	}
	future := c.handler.Raft.RemoveServer(raft.ServerID(req.Id), 0, 0)
	if err = future.Error(); err != nil {
		FAILED(ctx, 500, errors.Wrap(err, "can not remove node").Error())
		return
	}
	SUCCESS(ctx, nil)
}

func (c *CoordController) PeerStats(ctx *gin.Context) {
	info := c.handler.Raft.Stats()
	SUCCESS(ctx, info)
}
