package main

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/examples/rpcsrv/services"
	_ "github.com/billyyoyo/microj/plugins/broker/nats"
	_ "github.com/billyyoyo/microj/plugins/config/etcd"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
	"github.com/billyyoyo/microj/server/rpc"
)

func main() {
	app.NewApplication().
		Init().
		WithServer(
			(&rpc.RpcServer{}).
				RegService(services.NewExampleService().RegRpc).
				RegService(services.NewHelloService().RegRpc).
				RegService(services.NewTestService().RegRpc)).
		Run()
}
