package main

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/examples/gatewaysrv/codes"
	"github.com/billyyoyo/microj/examples/rpcsrv/proto"
	"github.com/billyyoyo/microj/logger"
	_ "github.com/billyyoyo/microj/plugins/broker/nats"
	_ "github.com/billyyoyo/microj/plugins/config/etcd"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
	"github.com/billyyoyo/microj/server/gateway"
	"github.com/billyyoyo/microj/util"
	"github.com/valyala/fasthttp"
	"net/http"
	"strings"
)

var (
	conf *GatewayConfig
)

type GatewayConfig struct {
	Routes    []gateway.Route `yaml:"route" mapstructure:"route"`
	WhiteList []string        `yaml:"white-list" mapstructure:"white-list"`
}

func (g *GatewayConfig) CanRefresh() bool {
	return true
}

func (g *GatewayConfig) KeyName() string {
	return "gateway"
}

func main() {
	app.NewApplication().
		Init(func() {
			var c GatewayConfig
			err := config.ScanWithRefresh(&c)
			if err != nil {
				logger.Fatal("gateway config load failed", err)
			}
			conf = &c
		}).
		WithServer(gateway.NewGateway().
			AddInterceptor(-11, fasthttp.CompressHandler).
			AddInterceptor(-10, gateway.RecoverHandler).
			AddInterceptor(-9, gateway.LogHandler).
			AddInterceptor(1, MatchHandler).
			AddInterceptor(3, AuthHandler).
			AddInterceptor(4, PermissionHandler).
			AddRpcEndpoint(proto.NewExampleGw()).
			AddRpcEndpoint(proto.NewTestGw()).
			AddRpcEndpoint(proto.NewHelloGw())).
		Run()

}

func MatchHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := util.Bytes2str(ctx.Path())
		for _, r := range conf.Routes {
			if strings.HasPrefix(path, r.Path) {
				ctx.Request.Header.Set(gateway.MICRO_SERVICE_NAME, r.Id)
				ctx.Request.Header.Set(gateway.MICRO_SERVICE_PATH, r.Path)
				ctx.Request.Header.Set(gateway.MICRO_SERVICE_SCHEMA, r.Schema)
				next(ctx)
				return
			}
		}
		ctx.Error("non router exist", http.StatusNotFound)
	}
}

func AuthHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		whitelist := []string{"/server-api/login"}
		originPath := util.Bytes2str(ctx.Request.URI().Path())
		for _, p := range whitelist {
			if p == originPath {
				next(ctx)
				return
			}
		}
		token := util.Bytes2str(ctx.Request.Header.Peek("Authorization"))
		if token == "" {
			gateway.RetFailed(ctx, codes.ERR_UNAUTHORIZATION, "unauthenticate")
			return
		}
		next(ctx)
	}
}

func PermissionHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		permissionlist := []string{"/edit"}
		path := util.Bytes2str(ctx.Path())
		token := util.Bytes2str(ctx.Request.Header.Peek("Authorization"))
		for _, p := range permissionlist {
			if p == path && token == "123" {
				gateway.RetFailed(ctx, codes.ERR_NO_PERMISSION, "no permission")
				return
			}
		}
		next(ctx)
	}
}
