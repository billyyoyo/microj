package main

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/examples/apisrv/controllers"
	_ "github.com/billyyoyo/microj/plugins/broker/nats"
	_ "github.com/billyyoyo/microj/plugins/config/etcd"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
	"github.com/billyyoyo/microj/server/api"
)

func main() {
	app.NewApplication().
		Init().
		WithServer((&api.ApiServer{}).RegController(controllers.NewExampleController().Apis())).
		Run()
}
