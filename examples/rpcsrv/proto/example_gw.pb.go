package proto

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/client"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/util"
)

type ExampleGw struct {
	cli     ExampleClient
	routers map[string]func(ctx context.Context, in []byte) (out interface{}, err error)
}

func (e *ExampleGw) LBName() string {
	return "server-rpc"
}

func (e *ExampleGw) ServiceName() string {
	return "exam"
}

func NewExampleGw() *ExampleGw {
	inst := &ExampleGw{
		routers: make(map[string]func(ctx context.Context, in []byte) (out interface{}, err error)),
	}
	return inst.Init()
}

func (e *ExampleGw) Do(ctx context.Context, method, path string, in []byte) (out []byte, err error) {
	if e.cli == nil {
		conn := client.NewRpcConn(e.LBName())
		e.cli = NewExampleClient(conn)
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

func (e *ExampleGw) Init() *ExampleGw {
	e.routers["GET:/server-rpc/example/call"] = e.Example_Call
	return e
}

func (e *ExampleGw) Example_Call(ctx context.Context, in []byte) (out interface{}, err error) {
	req := &Request{}
	err = util.QueryUnmarshal(in, req)
	if err != nil {
		err = errs.Wrap(errs.ERRCODE_GATEWAY, err.Error(), err)
		return
	}
	out, err = e.cli.Call(ctx, req)
	return
}
