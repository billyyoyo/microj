package gateway

import (
	"context"
	"errors"
	"fmt"
	"github.com/billyyoyo/microj/app"
	"github.com/billyyoyo/microj/config"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/registry"
	"github.com/billyyoyo/microj/util"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MICRO_SERVICE_NAME   = "micro-service-name"
	MICRO_SERVICE_PATH   = "micro-service-path"
	MICRO_SERVICE_SCHEMA = "micro-service-schema"

	CONTENT_TYPE = "application/json"
)

var (
	lbServiceRegexp *regexp.Regexp
)

func init() {
	lbServiceRegexp, _ = regexp.Compile("^/(.)+/(.)+/")
}

func NewGateway() *GatewayServer {
	g := &GatewayServer{}
	g.Init()
	return g
}

type GatewayServer struct {
	incep       interceptors
	r           *rand.Rand
	fastWatcher chan bool
	clients     []*fasthttp.HostClient
	doers       map[string]Doer
	mu          sync.RWMutex
	nextCli     nextClient
	ip          string
	s           *fasthttp.Server
	timeout     int64
}

type Route struct {
	Id     string `yaml:"id"`
	Path   string `yaml:"path"`
	Schema string `yaml:"schema"`
}

func (r *Route) String() string {
	return fmt.Sprintf("%s : %s :// %s", r.Id, r.Schema, r.Path)
}

type nextClient func(sn string) (*fasthttp.HostClient, error)

func (s *GatewayServer) Init() {
	s.r = rand.New(rand.NewSource(99))
	config.SetDefault("gateway.timeout", 30)
	s.timeout = config.GetInt64("gateway.timeout")
	s.fastWatcher = make(chan bool, 32)
	s.doers = make(map[string]Doer)
	registry.AddWatcher(s.fastWatcher)
	s.ip = util.GetIP()
	go s.watch()
}

func (s *GatewayServer) Run() {
	sort.Sort(s.incep)
	s.s = &fasthttp.Server{
		Handler: s.combineHandler(),
	}
	if err := s.s.ListenAndServe(fmt.Sprintf("%s:%d", app.Addr(), app.Port())); err != nil {
		logger.Fatal("gateway startup failed", err)
	}
}

func (s *GatewayServer) Stop() {
	s.s.Shutdown()
}

type interceptors []*interceptor

type interceptor struct {
	index int32
	next  func(next fasthttp.RequestHandler) fasthttp.RequestHandler
}

func (o interceptors) Len() int {
	return len(o)
}

func (o interceptors) Less(i, j int) bool {
	return o[i].index > o[j].index
}

func (o interceptors) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (s *GatewayServer) AddInterceptor(i int32, n func(next fasthttp.RequestHandler) fasthttp.RequestHandler) *GatewayServer {
	s.incep = append(s.incep, &interceptor{
		index: i,
		next:  n,
	})
	return s
}

func (s *GatewayServer) AddRpcEndpoint(d Doer) *GatewayServer {
	key := fmt.Sprintf("/%s/%s/", d.LBName(), strings.ToLower(d.ServiceName()))
	s.doers[key] = d
	return s
}

func (s *GatewayServer) combineHandler() fasthttp.RequestHandler {
	h := s.exec
	for _, inp := range s.incep {
		h = inp.next(h)
	}
	return h
}

func (s *GatewayServer) exec(ctx *fasthttp.RequestCtx) {
	req := &ctx.Request
	resp := &ctx.Response
	req.Header.Add("X-Forwarded-For", s.ip)
	serviceName := util.Bytes2str(req.Header.Peek(MICRO_SERVICE_NAME))
	servicePath := util.Bytes2str(req.Header.Peek(MICRO_SERVICE_PATH))
	serviceSchema := util.Bytes2str(req.Header.Peek(MICRO_SERVICE_SCHEMA))
	method := strings.ToUpper(util.Bytes2str(ctx.Method()))
	path := util.Bytes2str(ctx.Path())
	subPath := fmt.Sprintf("%s:///%s", serviceSchema, strings.ReplaceAll(path, servicePath, ""))
	if serviceSchema == "controller" {
		cli, err := s.selectCli(serviceName)
		if err != nil {
			logger.Error("remote controller call error", errors.New("no service instance"))
			body, _ := app.FailedResult(errs.ERRCODE_GATEWAY, "no service instance").Marshal()
			ctx.Success(CONTENT_TYPE, body)
			return
		}
		req.Header.SetHost(cli.Addr)
		req.Header.Del(MICRO_SERVICE_PATH)
		req.Header.Del(MICRO_SERVICE_NAME)
		ctx.Request.SetRequestURI(subPath)
		req.SetHost(cli.Addr)
		if err := cli.DoTimeout(req, resp, time.Duration(s.timeout)*time.Second); err != nil {
			logger.Error("remote controller call error", err)
			body, _ := app.FailedResult(errs.ERRCODE_GATEWAY, err.Error()).Marshal()
			ctx.Success(CONTENT_TYPE, body)
		}
	} else if serviceSchema == "rpc" {
		key := lbServiceRegexp.FindString(path)
		if key == "" {
			logger.Error("url parse error", errors.New("url parse error"))
			body, _ := app.FailedResult(errs.ERRCODE_GATEWAY, "url parse error").Marshal()
			ctx.Success(CONTENT_TYPE, body)
			return
		}
		if d, ok := s.doers[key]; ok {
			token := ctx.Request.Header.Peek("Authorization")
			c, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
			pair := metadata.Pairs("Authorization", util.Bytes2str(token))
			c = metadata.NewOutgoingContext(c, pair)
			defer cancel()
			var in []byte
			switch method {
			case http.MethodGet:
				in = req.URI().FullURI()
			case http.MethodPost:
				in = req.Body()
			case http.MethodPut:
				in = req.Body()
			case http.MethodDelete:
				in = req.URI().FullURI()
			default:
				logger.Error("method not allowed", errors.New("method not allowed"))
				body, _ := app.FailedResult(errs.ERRCODE_GATEWAY, "method not allowed").Marshal()
				ctx.Success(CONTENT_TYPE, body)
			}
			out, err := d.Do(c, method, path, in)
			resp.Header.Set("Content-Type", "application/json")
			resp.Header.Set("Date", time.Now().Format(time.RFC1123))
			if err != nil {
				var body []byte
				if st, ok := status.FromError(err); ok {
					body, _ = app.FailedResult(int(st.Code()), st.Message()).Marshal()
				} else {
					body, _ = app.FailedResult(errs.ERRCODE_GATEWAY, err.Error()).Marshal()
				}
				resp.Header.Set("Content-Length", strconv.Itoa(len(body)))
				logger.Error(err.Error(), err)
				ctx.Success(CONTENT_TYPE, body)
			} else {
				resp.Header.Set("Content-Length", strconv.Itoa(len(out)))
				resp.SetBody(out)
			}
			return
		} else {
			logger.Error("remote controller call error", errors.New("no endpoint instance"))
			body, _ := app.FailedResult(errs.ERRCODE_GATEWAY, "no endpoint instance").Marshal()
			ctx.Success(CONTENT_TYPE, body)
			return
		}
	} else {
		logger.Error("remote controller call error", errors.New("no support schema"))
		body, _ := app.FailedResult(errs.ERRCODE_GATEWAY, "no support schema").Marshal()
		ctx.Success(CONTENT_TYPE, body)
	}
}

func (s *GatewayServer) watch() {
	for {
		select {
		case <-s.fastWatcher:
			s.roundRobin()
		}
	}
}

func (s *GatewayServer) roundRobin() {
	var before, after []string
	mt := make(map[string]string) // key: addr value: serviceName
	for _, b := range s.clients {
		before = append(before, b.Addr)
	}
	serviceNames, _ := registry.ListServices()
	for i := 0; i < len(serviceNames); i++ {
		s := serviceNames[i]
		service, err := registry.GetService(s)
		if err != nil {
			break
		}
		for j := 0; j < len(service.Nodes); j++ {
			b := service.Nodes[j]
			after = append(after, b.Addr())
			mt[b.Addr()] = s
		}
	}
	adds := util.DeleteSlice(after, before)
	dels := util.DeleteSlice(before, after)
	s.mu.Lock()
	for i := 0; i < len(s.clients); i++ {
		cli := s.clients[i]
		if exist(dels, cli.Addr) {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
		}
	}
	for _, a := range adds {
		n := mt[a]
		s.clients = append(s.clients, &fasthttp.HostClient{
			Addr:      a,
			Name:      n,
			IsTLS:     false,
			TLSConfig: nil,
		})
	}
	s.mu.Unlock()

	var i = s.r.Int()
	s.nextCli = func(sn string) (*fasthttp.HostClient, error) {
		var clis []*fasthttp.HostClient
		for _, cli := range s.clients {
			if sn == cli.Name {
				clis = append(clis, cli)
			}
		}
		if len(clis) == 0 {
			return nil, errors.New("no available nodes")
		}
		s.mu.RLock()
		node := clis[i%len(clis)]
		i++
		s.mu.RUnlock()
		return node, nil
	}
}
func (s *GatewayServer) selectCli(serviceName string) (n *fasthttp.HostClient, err error) {
	if s.nextCli != nil {
		return s.nextCli(serviceName)
	}
	return nil, errs.New(errs.ERRCODE_GATEWAY, "service registry not ready")
}

func LogHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		start := time.Now()
		next(ctx)
		end := time.Now()
		latencyTime := end.Sub(start)
		reqMethod := ctx.Method()
		reqUri := util.Bytes2str(ctx.Request.RequestURI())
		statusCode := ctx.Response.StatusCode()
		clientIP := ctx.RemoteIP().String()
		logger.Infof("code=%d took=%dms ip=%s method=%s path=%s",
			statusCode,
			latencyTime.Milliseconds(),
			clientIP,
			reqMethod,
			reqUri,
		)
	}
}

func RecoverHandler(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if err := recover(); err != nil {
				req := ctx.Request.String()
				if ne, ok := err.(error); ok {
					ctx.Error(ne.Error(), http.StatusInternalServerError)
					logger.Error("Recover from panic",
						err.(error),
						logger.Val{K: "request", V: req},
					)
				} else {
					ctx.Error("unknown error", http.StatusInternalServerError)
					logger.Error("Recover from panic",
						nil,
						logger.Val{K: "request", V: req},
						logger.Val{K: "error", V: err},
					)
				}
			}
		}()
		next(ctx)
	}
}

func RetFailed(ctx *fasthttp.RequestCtx, code int, msg ...string) {
	body, _ := app.FailedResult(code, msg...).Marshal()
	ctx.Success(CONTENT_TYPE, body)
}

func exist(arr []string, s string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

type Doer interface {
	LBName() string
	ServiceName() string
	Do(ctx context.Context, method, path string, in []byte) (out []byte, err error)
}
