//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"user-service/internal/biz"
	"user-service/internal/conf"
	"user-service/internal/data"
	"user-service/internal/pkg"
	"user-service/internal/server"
	"user-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.JWT, *conf.IDGen, log.Logger, string) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, pkg.ProviderSet, newApp))
}
