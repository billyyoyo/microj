package nats

import (
	"fmt"
	"github.com/billyyoyo/microj/broker"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/logger"
	"github.com/nats-io/nats.go"
	"strings"
	"sync"
	"time"
)

type natsBroker struct {
	conn          *nats.Conn
	client        *nats.EncodedConn
	addr          string
	user          string
	pwd           string
	topicHandlers []topicHandler
	lock          sync.Mutex
}
type topicHandler struct {
	once    bool
	topic   string
	group   string
	handler broker.Handler
}

func init() {
	broker.InvokeInitBroker = newBroker
}

func newBroker(opts broker.Options) (broker.Broker, error) {
	if opts.Enable {
		b := &natsBroker{
			addr:   opts.Host,
			user:   opts.User,
			pwd:    opts.Pwd,
			client: nil,
		}
		return b, nil
	}
	return nil, nil
}

func (n *natsBroker) Init(opts broker.Options) error {
	if len(n.addr) == 0 {
		return errs.New(errs.ERRCODE_BROKER, "broker addr is empty")
	}
	addrs := strings.Split(n.addr, ",")
	for i, addr := range addrs {
		addrs[i] = "nats://" + addr
	}
	n.addr = strings.Join(addrs, ",")
	return nil
}

func (n *natsBroker) Connect() error {
	var options []nats.Option
	if n.user != "" && n.pwd != "" {
		options = append(options, nats.UserInfo(n.user, n.pwd))
	}
	options = append(options,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(1_000_000),
		nats.ReconnectWait(5*time.Second),
		nats.DisconnectErrHandler(n.onDisconnectError),
		nats.ErrorHandler(n.onError),
		nats.ConnectHandler(n.onConnect),
		nats.ReconnectHandler(n.onReconnect),
		nats.ClosedHandler(n.onClose))
	var err error
	n.conn, err = nats.Connect(n.addr, options...)
	if err != nil {
		return errs.Wrap(errs.ERRCODE_BROKER, err.Error(), err)
	}
	n.client, err = nats.NewEncodedConn(n.conn, nats.JSON_ENCODER)
	if err != nil {
		return errs.Wrap(errs.ERRCODE_BROKER, err.Error(), err)
	}
	return nil
}

func (n *natsBroker) Disconnect() error {
	n.client.Drain()
	n.client.Close()
	return nil
}

func (n *natsBroker) Receive(once bool, topic, group string, handler broker.Handler) error {
	if n.conn.IsConnected() {
		return n.receive(once, topic, group, handler)
	} else {
		n.topicHandlers = append(n.topicHandlers, topicHandler{
			once:    once,
			topic:   topic,
			group:   group,
			handler: handler,
		})
		return nil
	}
}

func (n *natsBroker) receive(once bool, topic, group string, handler broker.Handler) error {
	n.lock.Lock()
	defer n.lock.Unlock()
	var err error
	if once {
		_, err = n.client.QueueSubscribe(topic, group, handler)
	} else {
		_, err = n.client.Subscribe(topic, handler)
	}
	if err != nil {
		logger.Error(fmt.Sprintf("%s listen topic %s failed", group, topic), err)
	} else {
		logger.Info(fmt.Sprintf("%s listen topic %s success", group, topic))
	}
	return err
}

func (n *natsBroker) Send(topic string, msg broker.Message) error {
	err := n.client.Publish(topic, msg)
	return errs.Wrap(errs.ERRCODE_BROKER, "", err)
}

func (n *natsBroker) onDisconnectError(nc *nats.Conn, err error) {
	logger.Error("nats client disconnect ", err)
}

func (n *natsBroker) onReconnect(nc *nats.Conn) {
	logger.Info("nats client reconnect ")
}

func (n *natsBroker) onClose(nc *nats.Conn) {
	logger.Info("nats client closed ")
}

func (n *natsBroker) onConnect(nc *nats.Conn) {
	nc.RTT()
	logger.Info("nats client connect success ")
	for i := 0; i < len(n.topicHandlers); i++ {
		h := n.topicHandlers[i]
		if err := n.receive(h.once, h.topic, h.group, h.handler); err == nil {
			n.topicHandlers = append(n.topicHandlers[:i], n.topicHandlers[i+1:]...)
		}
	}
}

func (n *natsBroker) onError(nc *nats.Conn, sub *nats.Subscription, err error) {
	logger.Error("", err)
}
