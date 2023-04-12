package registry

import (
	"fmt"
	"github.com/billyyoyo/microj/logger"
)

const (
	SERVICE_PREFIX string = "/servers/"
)

var (
	ServiceRegistry    *Registry
	InvokeInitRegistry func(opts Options) (Registry, error)
	watchers           []chan bool
)

type Registry interface {
	Init(Options) error
	Register() error
	Deregister() error
	GetService(name string) (Service, error)
	ListServices() ([]string, error)
}

type Options struct {
	Host        string `yaml:"host"`
	User        string `yaml:"user"`
	Pwd         string `yaml:"pwd"`
	Ip          string `yaml:"ip"`
	ServiceName string `yaml:"serviceName"`
	Weight      int    `yaml:"weight"`
	Enable      bool   `yaml:"enable"`
	Port        int    `yaml:"port"`
	Timeout     int64  `yaml:"timeout"`
}

type Service struct {
	Name  string  `json:"name"`
	Nodes []*Node `json:"nodes"`
}

type Node struct {
	Id          string            `json:"id"`
	ServiceName string            `json:"serviceName"`
	Path        string            `json:"path"`
	Ip          string            `json:"ip"`
	Port        int               `json:"port"`
	Info        map[string]string `json:"info"`
}

func (n *Node) String() string {
	return fmt.Sprintf("%s\t%s:%d in %s", n.Id, n.Ip, n.Port, n.Path)
}

func (n *Node) Addr() string {
	return fmt.Sprintf("%s:%d", n.Ip, n.Port)
}

func Init(opts Options) {
	// todo 指向接口的指针，本不这样推荐使用，正确做法是把数据缓存部分单独创建结构体
	ServiceRegistry = new(Registry)
	reger, err := InvokeInitRegistry(opts)
	if err != nil {
		logger.Error("init error", err)
		return
	}
	ServiceRegistry = &reger
	Register()
	if err != nil {
		logger.Error("register error", err)
		return
	}
}

func Register() error {
	return (*ServiceRegistry).Register()
}

func Deregister() error {
	return (*ServiceRegistry).Deregister()
}

func GetService(name string) (Service, error) {
	return (*ServiceRegistry).GetService(name)
}

func ListServices() ([]string, error) {
	return (*ServiceRegistry).ListServices()
}

func AddWatcher(watcher chan bool) {
	watchers = append(watchers, watcher)
}

func NotifyWatcher() {
	for _, ch := range watchers {
		ch <- true
	}
}
