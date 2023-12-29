package main

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/examples/coordinatesrv/controller"
	"github.com/billyyoyo/microj/examples/coordinatesrv/coor"
	"github.com/billyyoyo/microj/server/api"
	"github.com/billyyoyo/microj/server/coordinate"
)

func main() {
	handler := &coor.CoordinatorHandler{}
	ctrl := controller.New(handler)
	app.NewApplication().Init().
		WithServer(coordinate.NewServer().WithHandler(handler)).
		WithServer((&api.ApiServer{}).RegFilter(ctrl.ProxyFilter).RegController(ctrl.Apis()...)).
		Run()
}
