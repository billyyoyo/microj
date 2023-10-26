package app

import (
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/broker"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/registry"
	"os"
	"os/signal"
	"strings"
)

var (
	app *application
)

type Server interface {
	Init()
	Run()
	Stop()
}

type application struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`

	servers []Server
}

func Name() string {
	return app.Name
}

func Addr() string {
	return app.Addr
}
func Port() int {
	return app.Port
}
func Mode() string {
	return app.Mode
}

func App() *application {
	return app
}

func (a *application) Init(fns ...func()) *application {
	config.Init()
	err := config.Scan("app", &app)
	if err != nil {
		logger.Fatal("app base info load failed", err)
	}
	if app.Addr == "" {
		app.Addr = "0.0.0.0"
	}
	a.initBack()
	if fns != nil {
		for _, fn := range fns {
			fn()
		}
	}
	return a
}

func (a *application) initBack() {
	//time.Sleep(5 * time.Second)
	var ro registry.Options
	err := config.Scan("registry", &ro)
	if err != nil {
		logger.Error("no mq broker", err)
		err = nil
	}
	ro.ServiceName = app.Name
	ro.Port = app.Port
	if ro.Host != "" {
		logger.Infof("registry address is %s:%s", ro.Host, ro.Port)
		registry.Init(ro)
	}
	bo := broker.Options{}
	err = config.Scan("broker", &bo)
	if err != nil {
		logger.Error("no mq broker", err)
		err = nil
	}
	if bo.Host != "" {
		logger.Infof("mq broker address is %s", bo.Host)
		broker.Init(bo)
	}
}

func (a *application) BrokerListener(f func()) *application {
	f()
	return a
}

func (a *application) WithServer(s Server) *application {
	a.servers = append(a.servers, s)
	return a
}

func (a *application) Run() {
	for _, s := range a.servers {
		s.Run()
	}
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	registry.Deregister()
	broker.Disconnect()
	logger.Info("Shutdown Server ...")
	for _, s := range a.servers {
		s.Stop()
	}
}

func NewApplication() *application {
	app = &application{}
	return app
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func SuccessResult(d any) *Result {
	return &Result{
		Code: 0,
		Data: d,
	}
}

func FailedResult(code int, msg ...string) *Result {
	return &Result{
		Code: code,
		Msg:  strings.Join(msg, " "),
	}
}

func FailedResultf(code int, format string, vals ...any) *Result {
	return &Result{
		Code: code,
		Msg:  fmt.Sprintf(format, vals...),
	}
}

func (r *Result) IsSuccess() bool {
	return r.Code == 0
}

func (r *Result) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
