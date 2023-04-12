package broker

import (
	"fmt"
	"github.com/billyyoyo/microj/util"
	"testing"
	"time"
)
import _ "github.com/billyyoyo/microj/plugins/broker/nats"

func TestNats(t *testing.T) {
	opts := Options{
		Enable: true,
		Addr:   "localhost:4222",
		User:   "",
		Pwd:    "",
	}
	Init(opts)
	Recv(false, "test", "nats", handleMsg)
	tick := time.Tick(time.Second)
	for t := range tick {
		text := t.Format(time.DateTime)
		msg := Message{
			Head: map[string]string{
				"token": "1111",
			},
			Body: []byte(text),
		}
		Send("test", msg)
	}
}

func handleMsg(msg Message) {
	fmt.Println(util.Bytes2str(msg.Body))
}
