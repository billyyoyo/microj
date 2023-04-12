package listeners

import (
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/broker"
	"github.com/billyyoyo/microj/logger"
)

type ExampleListener struct{}

func NewExampleListener() *ExampleListener {
	return &ExampleListener{}
}

func (h *ExampleListener) RegBroker() {
	broker.Recv(true, "example-hello", app.Name(), h.recvMsg)
}

func (h *ExampleListener) recvMsg(msg broker.Message) {
	logger.Info(msg)
}
