package listeners

import (
	"github.com/billyyoyo/microj/broker"
	"github.com/billyyoyo/microj/logger"
)

type ExampleListener struct{}

func NewExampleListener() *ExampleListener {
	return &ExampleListener{}
}

func (h *ExampleListener) RegBroker() {
	broker.Subscribe("example-hello", h.recvMsg)
	broker.Consume("example-hello", h.recvMsg)
}

func (h *ExampleListener) recvMsg(msg broker.Message) {
	logger.Info(msg)
}
