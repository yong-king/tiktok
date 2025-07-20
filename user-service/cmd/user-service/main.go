package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
	"user-service/internal/pkg"
	"user-service/internal/pkg/metrics"
	"user-service/internal/pkg/otelsetup"

	"user-service/internal/conf"

	consul "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hashicorp/consul/api"

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

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, consulAddr string) *kratos.App {
	// new consul client
	consulCfg := api.DefaultConfig()
	consulCfg.Address = consulAddr
	client, err := api.NewClient(consulCfg)
	if err != nil {
		panic(err)
	}
	reg := consul.New(client)

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

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		fmt.Println(err)
		panic(err)
	}

	var logCfg pkg.LogConfig
	if err := c.Scan(&logCfg); err != nil {
		panic(err)
	}
	logger := log.With(pkg.NewZapLogger(logCfg),
		//"ts", log.DefaultTimestamp,
		//"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	log.Debugf("bootstrap: %+v", bc)

	consulAddr := bc.Registry.Consul.Addr
	Name = bc.Service.Name
	Version = bc.Service.Version
	id = fmt.Sprintf("%s-%s", Name, bc.Server.Http.Addr)

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

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Jwt, bc.IdGen, logger, consulAddr)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// prometheus 监控
	metrics.Init()

	metrics.StartMetricsServer()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
