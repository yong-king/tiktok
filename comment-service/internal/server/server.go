package server

import (
	"comment-service/internal/conf"
	middleware "comment-service/internal/pkg/middle"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistrar)

// 创建限流中间件，限制全局入口请求
var RateLimitMw = middleware.RateLimitMiddleware(1000, 1888) // 100 QPS，突发200

func NewRegistrar(cfg *conf.Registry) registry.Registrar {
	// new consul client
	c := api.DefaultConfig()
	c.Address = cfg.Consul.Addr
	c.Scheme = cfg.Consul.Scheme
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// new reg with consul client
	reg := consul.New(client, consul.WithHealthCheck(true))
	return reg
}
