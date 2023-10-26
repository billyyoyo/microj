package handler

import (
	"github.com/billyyoyo/microj/examples/wssrv/protos"
	"github.com/billyyoyo/microj/logger"
	"github.com/golang/protobuf/proto"
	"github.com/lxzan/gws"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"
	"net/http"
)

type WSHandler struct {
	gws.BuiltinEventHandler
}

func (h *WSHandler) OnOpen(conn *gws.Conn) {
	name, ok := conn.Session().Load("name")
	if ok {
		logger.Info("user", name, "connected")
	} else {
		logger.Err(errors.New("no name in session"))
		conn.WriteClose(1000, []byte("no name in session"))
	}
}

func (h *WSHandler) OnClose(conn *gws.Conn, err error) {
	name, ok := conn.Session().Load("name")
	if ok {
		logger.Info("user", name, "disconnected")
	}
}

func (h *WSHandler) OnMessage(conn *gws.Conn, msg *gws.Message) {
	defer msg.Close()
	in := &protos.Pack{}
	if err := proto.Unmarshal(msg.Data.Bytes(), in); err == nil {
		out := &protos.Pack{}
		if in.Type == protos.PackType_PING {
			out.Type = protos.PackType_PONG
			out.Body, _ = anypb.New(&protos.Null{})
		} else if in.Type == protos.PackType_UP {
			name, _ := conn.Session().Load("name")
			inMsg := &protos.Message{}
			in.Body.UnmarshalTo(inMsg)
			out.Type = protos.PackType_DOWN
			out.Body, _ = anypb.New(&protos.Message{
				From:    "server",
				To:      name.(string),
				Content: inMsg.Content + "!!!",
			})
		}
		b, _ := proto.Marshal(out)
		conn.WriteMessage(gws.OpcodeBinary, b)

	} else {
		logger.Error("pack parse error: ", err)
	}

}

func (h *WSHandler) OnAuth(r *http.Request, session gws.SessionStorage) bool {
	var name = r.URL.Query().Get("name")
	var device = r.URL.Query().Get("device")
	if name == "admin" {
		session.Store("name", name)
		session.Store("device", device)
		return true
	}
	logger.Infof("auth check failed")
	return false
}
