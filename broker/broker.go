package broker

import (
	"encoding/json"
	"github.com/billyyoyo/microj/logger"
)

var (
	MqBroker         *Broker
	InvokeInitBroker func(opts Options) (Broker, error)
)

type Broker interface {
	Init(opts Options) error
	Connect() error
	Disconnect() error
	Receive(once bool, topic, group string, handler Handler) error
	Send(topic string, msg Message) error
}

type Options struct {
	Enable bool   `yaml:"enable"`
	Host   string `yaml:"host"`
	User   string `yaml:"user"`
	Pwd    string `yaml:"pwd"`
}

type Handler func(msg Message)

type Message struct {
	Head map[string]string `json:"head"`
	Body []byte            `json:"body"`
}

func (m Message) String() string {
	bs, _ := json.Marshal(m.Head)
	bs = append(bs, []byte(" | ")...)
	bs = append(bs, m.Body...)
	return string(bs)
}

func Init(opts Options) {
	MqBroker = new(Broker)
	b, err := InvokeInitBroker(opts)
	if err != nil {
		logger.Error("init broker error: ", err)
		return
	}
	MqBroker = &b
	err = Connect()
	if err != nil {
		logger.Error("connect broker error: ", err)
		return
	}
}

func Connect() error {
	return (*MqBroker).Connect()
}

func Disconnect() error {
	if MqBroker == nil {
		return nil
	}
	return (*MqBroker).Disconnect()
}

func Recv(once bool, topic, group string, handler Handler) error {
	err := (*MqBroker).Receive(once, topic, group, handler)
	if err == nil {
		logger.Infof("start listen mq %s at %s %t", topic, group, once)
	}
	return err
}

// Subscribe 一条消息被每个监听者都消费一次
func Subscribe(topic, group string, handler Handler) error {
	return Recv(false, topic, group, handler)
}

// Consume 一条消息被消费一次
func Consume(topic, group string, handler Handler) error {
	return Recv(true, topic, group, handler)
}

func Send(topic string, msg Message) {
	(*MqBroker).Send(topic, msg)
}
