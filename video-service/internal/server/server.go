package server

import (
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	"video-service/internal/conf"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewRegistry)

func NewRegistry(cfg *conf.Registry) registry.Registrar {
	c := api.DefaultConfig()
	c.Address = cfg.Consul.Addr
	c.Scheme = cfg.Consul.Scheme
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}

	reg := consul.New(client, consul.WithHealthCheck(true))
	return reg
}
