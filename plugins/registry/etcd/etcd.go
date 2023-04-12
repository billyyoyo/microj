package etcd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/registry"
	"github.com/billyyoyo/microj/util"
	clientv3 "go.etcd.io/etcd/client/v3"
	"path"
	"strings"
	"sync"
	"time"
)

var (
	MAX_KEEPALIVE_RETRY int8 = 3
)

type etcdRegistry struct {
	client         *clientv3.Client
	api            clientv3.KV
	options        registry.Options
	mu             sync.Mutex
	lease          clientv3.Lease
	leaseId        clientv3.LeaseID
	timeout        time.Duration
	self           registry.Node
	services       map[string]registry.Service
	breakCh        chan bool
	lastSyncTime   int64
	watchCancel    func()
	keepaliveRetry int8
}

func init() {
	registry.InvokeInitRegistry = NewRegistry
}

func NewRegistry(opts registry.Options) (r registry.Registry, err error) {
	if opts.Enable {
		ip := util.IF(opts.Ip != "", opts.Ip, util.GetIP()).(string)
		nodeId := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%d", ip, opts.Port))))
		r = &etcdRegistry{
			options: opts,
			self: registry.Node{
				Id:          nodeId,
				Path:        path.Join(registry.SERVICE_PREFIX, opts.ServiceName, nodeId),
				Ip:          ip,
				Port:        opts.Port,
				ServiceName: opts.ServiceName,
			},
			breakCh:        make(chan bool),
			keepaliveRetry: 3,
		}
		if err = r.Init(opts); err != nil {
			return
		}
	}
	return
}

func (e *etcdRegistry) Init(opts registry.Options) error {
	if opts.Host == "" {
		return errs.NewInternal("no registry servers address")
	}
	endpoints := strings.Split(opts.Host, ",")
	if e.client == nil {
		e.mu.Lock()
		defer e.mu.Unlock()
		newClient, err := clientv3.New(clientv3.Config{
			Endpoints: endpoints,
			Username:  opts.User,
			Password:  opts.Pwd,
		})
		if err != nil {
			return errs.Wrap(errs.ERRCODE_REGISTRY, "creating new etcd client for crypt.backend.Client", err)
		}
		e.client = newClient
		e.api = clientv3.NewKV(newClient)
		e.timeout = time.Second * 3
		return nil
	}
	return nil
}

func (e *etcdRegistry) Register() error {
	go e.loop()
	go e.watch()
	return nil
}

func (e *etcdRegistry) loop() {
	for {
		select {
		case <-e.breakCh:
			e.services = nil
			e.client.Delete(context.Background(), e.self.Path)
			e.leaseId = 0
			e.watchCancel()
			return
		default:
			time.Sleep(time.Second)
		}
		if e.leaseId == 0 {
			err := e.reg()
			if err != nil {
				logger.Error("register this service failed:", err)
				time.Sleep(5 * time.Second)
				continue
			}
		}
		if e.services == nil || time.Now().Unix()-e.lastSyncTime > 60 {
			err := e.allServices(registry.SERVICE_PREFIX)
			if err != nil {
				logger.Error("sync services list in cache failed:", err)
				e.services = nil
			}
		}
		kCtx, _ := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := e.lease.KeepAliveOnce(kCtx, e.leaseId)
		if err != nil {
			e.keepaliveRetry++
			logger.Info("lease keep alive failed", err.Error(), "retry: ", e.keepaliveRetry)
			if e.keepaliveRetry >= MAX_KEEPALIVE_RETRY {
				e.leaseId = 0
				e.mu.Lock()
				logger.Info("release all services cache")
				e.services = nil
				e.mu.Unlock()
			}
			continue
		} else {
			e.keepaliveRetry = 0
		}
	}
}

func (e *etcdRegistry) allServices(prefix string) error {
	existCtx, existCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer existCancel()
	resp, err := e.api.Get(existCtx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	services := make(map[string]registry.Service)
	for _, kv := range resp.Kvs {
		path := util.Bytes2str(kv.Key)
		path, _ = strings.CutPrefix(path, prefix)
		serviceName := path[0:strings.LastIndex(path, "/")]
		service, ok := services[serviceName]
		if !ok {
			nodes := new([]*registry.Node)
			service = registry.Service{
				Name:  serviceName,
				Nodes: *nodes,
			}
		}
		node := registry.Node{}
		err = json.Unmarshal(kv.Value, &node)
		if err != nil {
			err = nil
			continue
		}
		service.Nodes = append(service.Nodes, &node)
		services[serviceName] = service
	}
	e.mu.Lock()
	e.services = services
	e.lastSyncTime = time.Now().Unix()
	registry.NotifyWatcher()
	e.mu.Unlock()
	return nil
}

func (e *etcdRegistry) reg() error {
	if e.lease == nil {
		e.lease = clientv3.NewLease(e.client)
	}
	gctx, gcancel := context.WithTimeout(context.Background(), e.timeout*time.Second)
	defer gcancel()
	// todo 设置了超时context也会阻塞在这里  fuck
	leaseResp, err := e.lease.Grant(gctx, 10)
	if err != nil {
		return err
	}
	e.leaseId = leaseResp.ID
	metadata, _ := json.Marshal(e.self)
	pctx, pcancel := context.WithTimeout(context.Background(), e.timeout*time.Second)
	defer pcancel()
	_, err = e.api.Put(pctx, e.self.Path, util.Bytes2str(metadata), clientv3.WithLease(e.leaseId))
	if err != nil {
		e.leaseId = 0
		return err
	}
	logger.Info(fmt.Sprintf("register success %s LeaseID: %x\n", e.self.String(), e.leaseId))
	return nil
}

func (e *etcdRegistry) Deregister() error {
	e.breakCh <- true
	return nil
}

func (e *etcdRegistry) GetService(name string) (s registry.Service, err error) {
	s, ok := e.services[name]
	if !ok {
		err = errors.New("service not exist")
	}
	return
}

func (e *etcdRegistry) ListServices() ([]string, error) {
	var names []string
	if e.services == nil {
		return names, errs.New(errs.ERRCODE_REGISTRY, "no services")
	}
	for k, _ := range e.services {
		names = append(names, k)
	}
	return names, nil
}

func (e *etcdRegistry) watch() {
	wctx, wc := context.WithCancel(context.Background())
	e.watchCancel = wc
	wch := e.client.Watch(wctx, registry.SERVICE_PREFIX, clientv3.WithPrefix())
	for {
		select {
		case resp := <-wch:
			for _, ev := range resp.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					n := registry.Node{}
					if err := json.Unmarshal(ev.Kv.Value, &n); err != nil {
						continue
					}
					e.putNode(n)
				case clientv3.EventTypeDelete:
					e.deleteNode(util.Bytes2str(ev.Kv.Key))
				}
			}
		case <-wctx.Done():
			logger.Info("watcher be canceled")
			return
		}
	}
}

func (e *etcdRegistry) deleteNode(key string) {
	if e.services == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	path, _ := strings.CutPrefix(key, registry.SERVICE_PREFIX)
	serviceName := path[0:strings.LastIndex(path, "/")]
	nodeId := path[strings.LastIndex(path, "/")+1:]
	logger.Info("del node: ", serviceName, nodeId)
	service, ok := e.services[serviceName]
	if !ok {
		return
	}
	for i := 0; i < len(service.Nodes); i++ {
		n := service.Nodes[i]
		if n.Id == nodeId {
			service.Nodes = append(service.Nodes[:i], service.Nodes[i+1:]...)
			i--
		}
	}
	if len(service.Nodes) == 0 {
		delete(e.services, serviceName)
	} else {
		e.services[serviceName] = service
	}
	registry.NotifyWatcher()
}

func (e *etcdRegistry) putNode(node registry.Node) {
	if e.services == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	path, _ := strings.CutPrefix(node.Path, registry.SERVICE_PREFIX)
	serviceName := path[0:strings.LastIndex(path, "/")]
	logger.Info("put node:", serviceName, node.Id)
	service, ok := e.services[serviceName]
	if !ok {
		service = registry.Service{
			Name: serviceName,
		}
	}
	service.Nodes = append(service.Nodes, &node)
	e.services[serviceName] = service
	registry.NotifyWatcher()
}
