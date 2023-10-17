package main

import (
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
	"net/http"
	"strings"
)

const (
	version            = "0.0.1"
	deprecationComment = "// Deprecated: Do not use."
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-gin %v\n", version)
		return
	}

	var flags flag.FlagSet

	options := protogen.Options{
		ParamFunc: flags.Set,
	}

	options.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_gw.pb.go"
	pkg := &GRPCPackage{}
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	pkg.Package = string(file.GoPackageName)

	for _, service := range file.Services {
		ss := parseService(gen, file, g, service)
		if ss != nil {
			pkg.Services = append(pkg.Services, ss)
		}
	}
	g.P(strings.ReplaceAll(pkg_tpl, "{{$PACKAGE}}", pkg.Package))
	g.P()
	for _, s := range pkg.Services {
		st := strings.ReplaceAll(service_tpl, "{{$SERVICE_NAME}}", s.Name)
		st = strings.ReplaceAll(st, "{{$LB_NAME}}", s.LBName)
		st = strings.ReplaceAll(st, "{{$ENDPOINT_NAME}}", s.EndpointName)
		g.P(st)
		g.P()
		g.P(strings.ReplaceAll(init_head_tpl, "{{$SERVICE_NAME}}", s.Name))
		for _, m := range s.Methods {
			mt := strings.ReplaceAll(init_body_tpl, "{{$HTTP_METHOD}}", m.HttpMethod)
			mt = strings.ReplaceAll(mt, "{{$HTTP_PATH}}", m.HttpPath)
			mt = strings.ReplaceAll(mt, "{{$SERVICE_NAME}}", s.Name)
			mt = strings.ReplaceAll(mt, "{{$METHOD_NAME}}", m.Name)
			g.P(mt)
		}
		g.P(init_tail_tpl)
		for _, m := range s.Methods {
			g.P()
			mt := strings.ReplaceAll(do_func_tpl, "{{$SERVICE_NAME}}", s.Name)
			mt = strings.ReplaceAll(mt, "{{$METHOD_NAME}}", m.Name)
			mt = strings.ReplaceAll(mt, "{{$IN_PARAM}}", m.InParam)
			mt = strings.ReplaceAll(mt, "{{$UNMARSHAL_FUNC}}", m.UnmashalFunc)
			g.P(mt)
		}
	}

	return g
}

func parseService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, s *protogen.Service) *GRPCService {
	service := &GRPCService{}
	service.Name = s.GoName
	comment := strings.Trim(string(s.Comments.Leading), " ")
	if strings.HasPrefix(comment, "gw:") {
		tmp := strings.Trim(comment[strings.Index(comment, ":")+1:], " ")
		tmp = strings.Trim(tmp, "\n")
		names := strings.Split(tmp, ":")
		service.LBName = names[0]
		service.EndpointName = names[1]
	} else {
		return nil
	}
	for _, method := range s.Methods {
		mm := parseMethod(g, service, method)
		if mm == nil {
			continue
		}
		service.Methods = append(service.Methods, mm)
	}
	return service
}

func parseMethod(g *protogen.GeneratedFile, s *GRPCService, m *protogen.Method) *GRPCMethod {
	method := &GRPCMethod{}
	method.Name = m.GoName
	method.InParam = string(m.Input.Desc.Name())
	method.OutParam = string(m.Output.Desc.Name())
	comment := strings.Trim(string(m.Comments.Leading), " ")
	if strings.HasPrefix(comment, "gw:") {
		method.HttpMethod = strings.ToUpper(comment[strings.Index(comment, ":")+1 : strings.Index(comment, "\"")])
		method.HttpMethod = strings.Trim(method.HttpMethod, " ")
		path := comment[strings.Index(comment, "\"")+1 : strings.LastIndex(comment, "\"")]
		method.HttpPath = fmt.Sprintf("/%s/%s%s", s.LBName, s.EndpointName, path)
		if method.HttpMethod == http.MethodPost || method.HttpMethod == http.MethodPut {
			method.UnmashalFunc = "json.Unmarshal"
		} else {
			method.UnmashalFunc = "util.QueryUnmarshal"
		}
		return method
	} else {
		return nil
	}
}

type GRPCPackage struct {
	Package  string
	Services []*GRPCService
}

type GRPCService struct {
	Name         string
	LBName       string
	EndpointName string
	Methods      []*GRPCMethod
}

type GRPCMethod struct {
	Name         string
	HttpMethod   string
	HttpPath     string
	UnmashalFunc string
	InParam      string
	OutParam     string
}

const (
	pkg_tpl = `package {{$PACKAGE}}

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/billyyoyo/microj/client"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/util"
)`
	service_tpl = `type {{$SERVICE_NAME}}Gw struct {
	cli     {{$SERVICE_NAME}}Client
	routers map[string]func(ctx context.Context, in []byte) (out interface{}, err error)
}

func (e *{{$SERVICE_NAME}}Gw) LBName() string {
	return "{{$LB_NAME}}"
}

func (e *{{$SERVICE_NAME}}Gw) ServiceName() string {
	return "{{$ENDPOINT_NAME}}"	
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
}`
	init_head_tpl = `func (e *{{$SERVICE_NAME}}Gw) Init() *{{$SERVICE_NAME}}Gw {`
	init_body_tpl = `    e.routers["{{$HTTP_METHOD}}:{{$HTTP_PATH}}"] = e.{{$SERVICE_NAME}}_{{$METHOD_NAME}}`
	init_tail_tpl = `    return e
}`
	// todo how to handle remote err
	do_func_tpl = `func (e *{{$SERVICE_NAME}}Gw) {{$SERVICE_NAME}}_{{$METHOD_NAME}}(ctx context.Context, in []byte) (out interface{}, err error) {
	req := &{{$IN_PARAM}}{}
	err = {{$UNMARSHAL_FUNC}}(in, req)
	if err != nil {
		err = errs.Wrap(errs.ERRCODE_GATEWAY, err.Error(), err)
		return
	}
	out, err = e.cli.{{$METHOD_NAME}}(ctx, req)
	return
}`
)
