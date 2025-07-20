package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/hashicorp/consul/api"
	"os"
	"os/signal"
	"syscall"
	"time"
	"video-service/internal/conf"
	"video-service/internal/pkg/otelsetup"
	"video-service/internal/server"
	"video-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, reg *consul.Registry) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.Registrar(reg),
	)
}

func newAppWithService(
	logger log.Logger,
	gs *grpc.Server,
	hs *http.Server,
	videoService *service.VideoService,
	reg *consul.Registry,
) (*kratos.App, func(), error) {
	// 绑定可供 Gin 使用的全局 VideoService
	service.BindVideoService(videoService)

	app := newApp(logger, gs, hs, reg)
	cleanup := func() {
		log.NewHelper(logger).Info("cleanup called")
	}
	return app, cleanup, nil
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	Name = bc.Service.Name
	Version = bc.Service.Version
	id = fmt.Sprintf("%s-%s", Name, bc.Server.Http.Addr)

	consulCfg := api.DefaultConfig()
	consulCfg.Address = bc.Registry.Consul.Addr
	client, err := api.NewClient(consulCfg)
	if err != nil {
		panic(err)
	}
	reg := consul.New(client)

	// ---------------OpenTelemetry--------------------
	ctx := context.Background()
	shutdown := otelsetup.InitTracerProvider(ctx, Name)
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger, bc.Jwt, bc.Data.Minio, bc.IdGen, reg, bc.Elasticsearch)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// 启动 Gin server
	go func() {
		r := server.NewGinServer()
		ginPort := bc.Server.Gin.Port
		addr := fmt.Sprintf(":%d", ginPort)
		if err := r.Run(addr); err != nil {
			panic(err)
		}
	}()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}

	// 优雅退出
	o := make(chan os.Signal, 1)
	signal.Notify(o, syscall.SIGINT, syscall.SIGTERM)
	<-o
}
