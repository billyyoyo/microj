package test

import (
	"github.com/billyyoyo/microj/broker"
	"testing"
	"time"
)
import _ "github.com/billyyoyo/microj/plugins/broker/nats"

func TestNats(t *testing.T) {
	opts := broker.Options{
		Enable: true,
		Host:   "localhost:4222",
		User:   "",
		Pwd:    "",
	}
	broker.Init(opts)
	//Recv(false, "test", "nats", handleMsg)
	tick := time.Tick(3 * time.Second)
	for t := range tick {
		text := t.Format(time.DateTime)
		msg := broker.Message{
			Head: map[string]string{
				"token": "1111",
			},
			Body: []byte(text),
		}
		broker.Send("example-hello", msg)
	}
}

//func handleMsg(msg Message) {
//	fmt.Println(util.Bytes2str(msg.Body))
//}
