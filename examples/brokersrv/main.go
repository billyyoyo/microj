package main

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/examples/brokersrv/listeners"
	_ "github.com/billyyoyo/microj/plugins/broker/nats"
	_ "github.com/billyyoyo/microj/plugins/config/etcd"
	_ "github.com/billyyoyo/microj/plugins/registry/etcd"
)

func main() {
	app.NewApplication().
		Init().
		BrokerListener(listeners.NewExampleListener().RegBroker).
		Run()
}
