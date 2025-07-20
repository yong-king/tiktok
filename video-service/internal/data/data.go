package data

import (
	"context"
	"errors"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"time"
	"video-service/internal/conf"
	"video-service/internal/data/query"
	"video-service/internal/pkg"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	pbUser "video-service/api/user/v1"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewVideoRepo, NewDB, NewRedisClient, NewEsClient)

// Data .
type Data struct {
	// TODO wrapped database client
	log     *log.Helper
	jwt     *pkg.JWTManager
	uploade *pkg.MinioUploader
	db      *gorm.DB
	rdb     *redis.Client
	idg     *pkg.IDGenerator
	query   *query.Query
	es      *elasticsearch.TypedClient
	esIndex string

	UserClient pbUser.UserServiceClient
}

// NewData .
func NewData(c *conf.Data, logger log.Logger, jwt *pkg.JWTManager, upload *pkg.MinioUploader, db *gorm.DB, rdb *redis.Client, idg *pkg.IDGenerator, rr *consul.Registry, esCfg *conf.Elasticsearch, es *elasticsearch.TypedClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	query.SetDefault(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(c.UserService.Endpoint),
		grpc.WithDiscovery(rr),
	)
	if err != nil {
		return nil, nil, err
	}
	return &Data{log: log.NewHelper(logger),
		jwt:     jwt,
		uploade: upload,
		db:      db, rdb: rdb,
		idg:        idg,
		query:      query.Q,
		UserClient: pbUser.NewUserServiceClient(conn),
		es:         es,
		esIndex:    esCfg.Index,
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
