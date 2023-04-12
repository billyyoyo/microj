package {{$PACKAGE}}

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/client"
	"github.com/billyyoyo/microj/util"
	"github.com/pkg/errors"
)

type {{$SERVICE_NAME}}Gw struct {
	cli     {{$SERVICE_NAME}}Client
	routers map[string]func(ctx context.Context, in []byte) (out interface{}, err error)
}

func (e *{{$SERVICE_NAME}}Gw) LBName() string {
	return "{{$LB_NAME}}"
}

func (e *{{$SERVICE_NAME}}Gw) ServiceName() string {
	return "{{$SERVICE_NAME}}"
}

func New{{$SERVICE_NAME}}Gw() *{{$SERVICE_NAME}}Gw {
	inst := &{{$SERVICE_NAME}}Gw{
		routers: make(map[string]func(ctx context.Context, in []byte) (out interface{}, err error)),
	}
	return inst.Init()
}

func (e *{{$SERVICE_NAME}}Gw) Do(ctx context.Context, method, path string, in []byte) (out []byte, err error) {
	if e.cli == nil {
		conn := client.NewRpcConn(e.LBName())
		e.cli = New{{$SERVICE_NAME}}Client(conn)
	}
	key := fmt.Sprintf("%s:%s", method, path)
	if fn, ok := e.routers[key]; ok {
		var o interface{}
		o, err = fn(ctx, in)
		out, err = json.Marshal(&o)
		return
	}
	err = errors.New("endpoint not found")
	return
}

func (e *{{$SERVICE_NAME}}Gw) Init() *{{$SERVICE_NAME}}Gw {
	e.routers["{{$HTTP_METHOD}}:{{$HTTP_PATH}}"] = e.{{$SERVICE_NAME}}_{{$METHOD_NAME}}
	return e
}



func (e *{{$SERVICE_NAME}}Gw) {{$SERVICE_NAME}}_{{$METHOD_NAME}}(ctx context.Context, in []byte) (out interface{}, err error) {
	req := &{{$IN_PARAM}}{}
	err = {{$UNMARSHAL_FUNC}}(in, req)
	if err != nil {
		return
	}
	out, err = e.cli.{{$METHOD_NAME}}(ctx, req)
	return
}
