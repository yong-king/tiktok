//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"favorite-service/internal/biz"
	"favorite-service/internal/conf"
	"favorite-service/internal/data"
	"favorite-service/internal/server"
	"favorite-service/internal/service"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger, *consul.Registry) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
