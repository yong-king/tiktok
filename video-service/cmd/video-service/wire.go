//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"video-service/internal/biz"
	"video-service/internal/conf"
	"video-service/internal/data"
	"video-service/internal/pkg"
	"video-service/internal/server"
	"video-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger, *conf.JWT, *conf.Data_MinIO, *conf.IDGen, *conf.Registry, *conf.Elasticsearch, *conf.OpenTelemetry) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, pkg.ProviderSet, newAppWithService))
}
