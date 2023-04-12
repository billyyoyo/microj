# Examples

4个示例

- **apisrv** http服务，使用gin cmd： go run -c api-srv-1.yml 
```
func main() {
	app.NewApplication().
		Init().
		WithServer((&api.ApiServer{}).RegController(controllers.NewExampleController().Apis())).
		Run()
}

func (c *ExampleController) Apis() []api.Api {
    return []api.Api{
        {Method: "post", Path: "/login", Func: c.Login},
        {Method: "get", Path: "/info", Func: c.GetInfo},
        {Method: "post", Path: "/edit", Func: c.Edit},
    }
}
```

- **brokersrv** 消息队列处理 cmd: go run -c rpc-srv-1.yml (一般不会这样)
```
func main() {
	app.NewApplication().
		Init().
		BrokerListener(listeners.NewExampleListener().RegBroker).
		Run()
}

func (h *ExampleListener) RegBroker() {
	broker.Recv(true, "example-hello", app.Name(), h.recvMsg)
}
```

- **rpcsrv** grpc服务 cmd：go run -c rpc-srv-1.yml  
```
func main() {
	app.NewApplication().
		Init().
		WithServer(
			(&rpc.RpcServer{}).
				RegService(services.NewExampleService().RegRpc).
				RegService(services.NewHelloService().RegRpc).
				RegService(services.NewTestService().RegRpc)).
		Run()
}

func (s *ExampleService) RegRpc(r *grpc.Server) {
	proto.RegisterExampleServer(r, NewExampleService())
}
```

- **gatewaysrv** 网关 cmd: go run -c gateway.yml
```
var (
	conf *GatewayConfig
)

type GatewayConfig struct {
	Routes    []gateway.Route `yaml:"route" mapstructure:"route"`
	WhiteList []string        `yaml:"white-list" mapstructure:"white-list"`
}

func (g *GatewayConfig) CanRefresh() bool {
return true
    }

func (g *GatewayConfig) KeyName() string {
return "gateway"
    }

func main() {
	app.NewApplication().
		Init(func() {
			var c GatewayConfig
			err := config.ScanWithRefresh(&c)
			if err != nil {
				logger.Fatal("gateway config load failed", err)
			}
			conf = &c
		}).
		WithServer(gateway.NewGateway().
			AddInterceptor(-11, fasthttp.CompressHandler).
			AddInterceptor(-10, gateway.RecoverHandler).
			AddInterceptor(-9, gateway.LogHandler).
			AddInterceptor(1, MatchHandler).
			AddInterceptor(3, AuthHandler).
			AddInterceptor(4, PermissionHandler).
			AddRpcEndpoint(proto.NewExampleGw()).
			AddRpcEndpoint(proto.NewTestGw()).
			AddRpcEndpoint(proto.NewHelloGw())).
		Run()

}
```