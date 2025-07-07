package data

import (
	"context"
	"errors"
	pbUSer "favorite-service/api/user/v1"
	pbVideo "favorite-service/api/video/v1"
	"favorite-service/internal/conf"
	"favorite-service/internal/data/query"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewFavoriteRepo, NewDB, NewRedisClient)

// Data .
type Data struct {
	// TODO wrapped database client
	log   *log.Helper
	db    *gorm.DB
	rdb   *redis.Client
	query *query.Query

	UserClient  pbUSer.UserServiceClient
	VideoClient pbVideo.VideoServiceClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, rr *consul.Registry) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)

	conn1, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint(c.UserService.Endpoint),
		grpc.WithDiscovery(rr))
	if err != nil {
		return nil, cleanup, err
	}
	conn2, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint(c.VideoService.Endpoint),
		grpc.WithDiscovery(rr))

	return &Data{log: log.NewHelper(logger), db: db, rdb: rdb, UserClient: pbUSer.NewUserServiceClient(conn1), query: query.Q, VideoClient: pbVideo.NewVideoServiceClient(conn2)}, cleanup, nil
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
