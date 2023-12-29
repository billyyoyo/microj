package client

import (
	"fmt"
	"github.com/billyyoyo/microj/errs"
	"github.com/billyyoyo/microj/logger"
	"github.com/billyyoyo/microj/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

var (
	builder *rpcResolverBuilder
)

func init() {
	builder = &rpcResolverBuilder{
		serviceWatcher: make(chan bool, 32),
	}
	registry.AddWatcher(builder.serviceWatcher)
	go builder.watch()
}

func NewRpcConn(serviceName string) *grpc.ClientConn {
	resolver.Register(builder)
	conn, err := grpc.Dial(
		fmt.Sprintf("lb:///%s", serviceName),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`), // This sets the initial balancing policy. could add health check config here: "healthCheckConfig": { "serviceName": "service-user"  }
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		err = errs.Wrap(errs.ERRCODE_REMOTE_CALL, err.Error(), err)
		logger.Fatal("proto conn did not connect: %v", err)
	}
	return conn
}

type rpcResolverBuilder struct {
	serviceWatcher chan bool
	resolvers      []*rpcResolver
}

func (b *rpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &rpcResolver{
		target: target,
		cc:     cc,
	}
	b.resolvers = append(b.resolvers, r)
	r.Notify()
	return r, nil
}

func (b *rpcResolverBuilder) Scheme() string { return "lb" }

func (b *rpcResolverBuilder) watch() {
	for {
		select {
		case <-b.serviceWatcher:
			for _, r := range b.resolvers {
				r.Notify()
			}
		}
	}
}

type rpcResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
}

func (*rpcResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (*rpcResolver) Close() {}

func (r *rpcResolver) Notify() {
	service, err := registry.GetService(r.target.Endpoint())
	if err == nil && len(service.Nodes) > 0 {
		var addrs = make([]resolver.Address, len(service.Nodes))
		for i, n := range service.Nodes {
			addrs[i] = resolver.Address{Addr: fmt.Sprintf("%s:%d", n.Ip, n.Port)}
		}
		r.cc.UpdateState(resolver.State{Addresses: addrs})
		return
	}
	r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{}})
}
