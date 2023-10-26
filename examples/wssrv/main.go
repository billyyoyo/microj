package main

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/examples/wssrv/handler"
	_ "github.com/billyyoyo/microj/plugins/config/etcd"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
	"github.com/billyyoyo/microj/server/ws"
)

func main() {
	h := &handler.WSHandler{}
	app.NewApplication().
		Init().
		WithServer((&ws.WSServer{}).
			RegAuth(h.OnAuth).
			RegEventHandler(h),
		).
		Run()
}
