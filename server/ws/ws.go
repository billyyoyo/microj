package ws

import (
	"fmt"
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/logger"
	"github.com/lxzan/gws"
	"net/http"
)

type AuthFilter func(r *http.Request, session gws.SessionStorage) bool
type OnOpenCallback func(socket *gws.Conn)
type OnCloseCallback func(socket *gws.Conn, err error)
type OnMsgCallback func(socket *gws.Conn, message *gws.Message)

type WSServer struct {
	gws.BuiltinEventHandler
	authFunc  AuthFilter
	wsHandler gws.Event
	app       *gws.Server
}

func (s *WSServer) Init() {
}

func (s *WSServer) Run() {
	s.app = gws.NewServer(s.wsHandler, &gws.ServerOption{
		CompressEnabled: true,
		Logger:          &WSLogger{},
		Authorize:       s.authFunc,
	})
	addr := fmt.Sprintf(":%d", app.Port())
	logger.Fatal("startup error", s.app.Run(addr))
}

func (s *WSServer) Stop() {
	logger.Info("ws server stop right now")
}

func (s *WSServer) RegAuth(filter AuthFilter) *WSServer {
	s.authFunc = filter
	return s
}

func (s *WSServer) RegEventHandler(h gws.Event) *WSServer {
	s.wsHandler = h
	return s
}

type WSLogger struct{}

func (l *WSLogger) Error(arr ...any) {
	if len(arr) > 2 {
		err := arr[1].(error)
		msg := arr[0].(string) + arr[2].(string)
		logger.Error(msg, err)
	}
	logger.Warn(arr...)
}
