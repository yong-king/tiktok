//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"comment-service/internal/biz"
	"comment-service/internal/conf"
	"comment-service/internal/data"
	"comment-service/internal/pkg"
	"comment-service/internal/server"
	"comment-service/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger, *conf.Registry, *conf.IDGen, *conf.OpenTelemetry) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, pkg.ProviderSet, newApp))
}
