package data

import (
	"comment-service/internal/conf"
	"comment-service/internal/data/query"
	"comment-service/internal/pkg"
	"context"
	"errors"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	gogrpc "google.golang.org/grpc" // 引入底层 grpc 包
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"

	pbUser "comment-service/api/user/v1"
	pbVideo "comment-service/api/video/v1"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCommentRepo, NewDB, NewRedisClient, NewDiscover, NewUserServiceClient, NewVideoServiceClient)

// Data .
type Data struct {
	// TODO wrapped database client
	log *log.Helper
	db  *gorm.DB
	rdb *redis.Client
	idg *pkg.IDGenerator

	query       *query.Query
	UserClient  pbUser.UserServiceClient
	VideoClient pbVideo.VideoServiceClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, cu pbUser.UserServiceClient, cv pbVideo.VideoServiceClient, idg *pkg.IDGenerator) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)

	return &Data{
		log:         log.NewHelper(logger),
		db:          db,
		rdb:         rdb,
		query:       query.Q,
		UserClient:  cu,
		VideoClient: cv,
		idg:         idg,
	}, cleanup, nil
}

func NewDiscover(cfg *conf.Registry) registry.Discovery {
	// new consul client
	c := api.DefaultConfig()
	c.Address = cfg.Consul.Addr
	c.Scheme = cfg.Consul.Scheme
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	// new dis with consul client
	reg := consul.New(client)
	return reg
}

func NewUserServiceClient(c *conf.Data, rr registry.Discovery) pbUser.UserServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(c.UserService.Endpoint),
		grpc.WithDiscovery(rr),
		grpc.WithOptions(gogrpc.WithStatsHandler(otelgrpc.NewClientHandler())),
	)
	if err != nil {
		panic(err)
	}
	return pbUser.NewUserServiceClient(conn)
}

func NewVideoServiceClient(c *conf.Data, rr registry.Discovery) pbVideo.VideoServiceClient {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(c.VideoService.Endpoint),
		grpc.WithDiscovery(rr),
		grpc.WithOptions(gogrpc.WithStatsHandler(otelgrpc.NewClientHandler())),
	)
	if err != nil {
		panic(err)
	}
	return pbVideo.NewVideoServiceClient(conn)
}

// NewDB 数据库连接
func NewDB(cfg *conf.Data) (*gorm.DB, error) {
	switch strings.ToLower(cfg.Database.Driver) {
	case "mysql":
		return gorm.Open(mysql.Open(cfg.Database.Source))
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.Database.Source))
	}
	return nil, errors.New("connect db failed unsuppoesd db driver")
}

// NewRedisClient 连接redis
func NewRedisClient(cfg *conf.Data) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		WriteTimeout: cfg.Redis.WriteTimeout.AsDuration(),
		ReadTimeout:  cfg.Redis.ReadTimeout.AsDuration(),
	})
}
