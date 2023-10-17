package proto

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/client"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/util"
)

type TestGw struct {
	cli     TestClient
	routers map[string]func(ctx context.Context, in []byte) (out interface{}, err error)
}

func (e *TestGw) LBName() string {
	return "server-rpc"
}

func (e *TestGw) ServiceName() string {
	return "test"
}

func NewTestGw() *TestGw {
	inst := &TestGw{
		routers: make(map[string]func(ctx context.Context, in []byte) (out interface{}, err error)),
	}
	return inst.Init()
}

func (e *TestGw) Do(ctx context.Context, method, path string, in []byte) (out []byte, err error) {
	if e.cli == nil {
		conn := client.NewRpcConn(e.LBName())
		e.cli = NewTestClient(conn)
	}
	key := fmt.Sprintf("%s:%s", method, path)
	if fn, ok := e.routers[key]; ok {
		if o, errr := fn(ctx, in); errr == nil {
			var er error
			out, er = json.Marshal(&o)
			if er != nil {
				err = er
			}
			return
		} else {
			err = errr
		}
		return
	}
	err = errs.NewInternal("endpoint not found")
	return
}

func (e *TestGw) Init() *TestGw {
	e.routers["GET:/server-rpc/test/exec"] = e.Test_Exec
	return e
}

func (e *TestGw) Test_Exec(ctx context.Context, in []byte) (out interface{}, err error) {
	req := &InParam{}
	err = util.QueryUnmarshal(in, req)
	if err != nil {
		err = errs.Wrap(errs.ERRCODE_GATEWAY, err.Error(), err)
		return
	}
	out, err = e.cli.Exec(ctx, req)
	return
}

type HelloGw struct {
	cli     HelloClient
	routers map[string]func(ctx context.Context, in []byte) (out interface{}, err error)
}

func (e *HelloGw) LBName() string {
	return "server-rpc"
}

func (e *HelloGw) ServiceName() string {
	return "hello"
}

func NewHelloGw() *HelloGw {
	inst := &HelloGw{
		routers: make(map[string]func(ctx context.Context, in []byte) (out interface{}, err error)),
	}
	return inst.Init()
}

func (e *HelloGw) Do(ctx context.Context, method, path string, in []byte) (out []byte, err error) {
	if e.cli == nil {
		conn := client.NewRpcConn(e.LBName())
		e.cli = NewHelloClient(conn)
	}
	key := fmt.Sprintf("%s:%s", method, path)
	if fn, ok := e.routers[key]; ok {
		if o, errr := fn(ctx, in); errr == nil {
			var er error
			out, er = json.Marshal(&o)
			if er != nil {
				err = er
			}
			return
		} else {
			err = errr
		}
		return
	}
	err = errs.NewInternal("endpoint not found")
	return
}

func (e *HelloGw) Init() *HelloGw {
	e.routers["POST:/server-rpc/hello/say"] = e.Hello_Say
	return e
}

func (e *HelloGw) Hello_Say(ctx context.Context, in []byte) (out interface{}, err error) {
	req := &InParam{}
	err = json.Unmarshal(in, req)
	if err != nil {
		err = errs.Wrap(errs.ERRCODE_GATEWAY, err.Error(), err)
		return
	}
	out, err = e.cli.Say(ctx, req)
	return
}
