package data

import (
	"context"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	pbUser "ralation-service/api/user/v1"
	"ralation-service/internal/conf"
	"ralation-service/internal/data/query"
	"strings"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewRelationRepo, NewDB, NewRedisClient, NewEsClient, NewDiscover, UserClient)

// Data .
type Data struct {
	// TODO wrapped database client
	log *log.Helper
	db  *gorm.DB
	rdb *redis.Client

	query   *query.Query
	es      *elasticsearch.TypedClient
	esIndex string

	UserClient pbUser.UserServiceClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, db *gorm.DB, rdb *redis.Client, es *elasticsearch.TypedClient, esCfg *conf.Elasticsearch, cu pbUser.UserServiceClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)

	return &Data{
		log:        log.NewHelper(logger),
		db:         db,
		rdb:        rdb,
		query:      query.Q,
		es:         es,
		esIndex:    esCfg.Index,
		UserClient: cu,
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

func NewEsClient(cfg *conf.Elasticsearch) (*elasticsearch.TypedClient, error) {
	c := elasticsearch.Config{
		Addresses: cfg.Addresses,
	}
	return elasticsearch.NewTypedClient(c)
}

func NewDiscover(cfg *conf.Registry) registry.Discovery {
	c := api.DefaultConfig()
	c.Address = cfg.Consul.Addr
	c.Scheme = "http"
	client, err := api.NewClient(c)
	if err != nil {
		panic(err)
	}
	reg := consul.New(client, consul.WithHealthCheck(true))
	return reg
}

func UserClient(c *conf.Data, rr registry.Discovery) pbUser.UserServiceClient {
	connUser, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(c.UserService.Endpoint),
		grpc.WithDiscovery(rr))
	if err != nil {
		panic(err)
	}
	return pbUser.NewUserServiceClient(connUser)
}
