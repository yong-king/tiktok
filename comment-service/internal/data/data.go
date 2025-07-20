package data

import (
	"comment-service/internal/conf"
	"comment-service/internal/data/query"
	"comment-service/internal/pkg"
	"context"
	"errors"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"gorm.io/gorm"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	gogrpc "google.golang.org/grpc" // 引入底层 grpc 包
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"

	pbUser "comment-service/api/user/v1"
	pbVideo "comment-service/api/video/v1"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewCommentRepo, NewDB, NewRedisClient)

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
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, rr *consul.Registry, idg *pkg.IDGenerator) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connUser, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(c.UserService.Endpoint),
		grpc.WithDiscovery(rr),
		grpc.WithOptions(gogrpc.WithStatsHandler(otelgrpc.NewClientHandler())),
	)
	if err != nil {
		return nil, cleanup, err
	}

	connVideo, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(c.VideoService.Endpoint),
		grpc.WithDiscovery(rr),
		grpc.WithOptions(gogrpc.WithStatsHandler(otelgrpc.NewClientHandler())),
	)
	if err != nil {
		return nil, cleanup, err
	}

	return &Data{
		log:         log.NewHelper(logger),
		db:          db,
		rdb:         rdb,
		query:       query.Q,
		UserClient:  pbUser.NewUserServiceClient(connUser),
		VideoClient: pbVideo.NewVideoServiceClient(connVideo),
		idg:         idg,
	}, cleanup, nil
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
