package client

import (
	"fmt"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/registry"
	"github.com/go-resty/resty/v2"
	"math/rand"
	"strings"
	"sync"
)

var (
	restWatcher chan bool
	r           = rand.New(rand.NewSource(99))
	next        Next
)

func init() {
	restWatcher = make(chan bool, 32)
	registry.AddWatcher(restWatcher)
	go watch()
}

func NewApiClient(serviceName string) *resty.Client {
	cli := resty.New()
	cli.SetBaseURL(fmt.Sprintf("lb://%s", serviceName))
	cli.OnBeforeRequest(onBefore)
	return cli
}

func onBefore(cli *resty.Client, req *resty.Request) error {
	if strings.HasPrefix(cli.BaseURL, "lb://") {
		serviceName := strings.TrimLeft(cli.BaseURL, "lb://")
		n, err := selectNode(serviceName)
		if err != nil {
			return errs.Wrap(errs.ERRCODE_REMOTE_CALL, err.Error(), err)
		}
		req.URL = fmt.Sprintf("http://%s:%d%s", n.Ip, n.Port, req.URL)
	}
	return nil
}

func watch() {
	for {
		select {
		case <-restWatcher:
			roundRobin()
		}
	}
}

type Next func(sn string) (*registry.Node, error)

func roundRobin() {
	var i = r.Int()
	var mtx sync.Mutex
	next = func(sn string) (*registry.Node, error) {
		var nodes []*registry.Node
		if service, err := registry.GetService(sn); err == nil {
			nodes = service.Nodes
		} else {
			return nil, errs.New(errs.ERRCODE_REMOTE_CALL, "service no exist")
		}
		if len(nodes) == 0 {
			return nil, errs.New(errs.ERRCODE_REMOTE_CALL, "no available nodes")
		}
		mtx.Lock()
		node := nodes[i%len(nodes)]
		i++
		mtx.Unlock()
		return node, nil
	}
}

func selectNode(serviceName string) (n *registry.Node, err error) {
	if next != nil {
		return next(serviceName)
	}
	return nil, errs.New(errs.ERRCODE_REMOTE_CALL, "service registry not ready")
}
