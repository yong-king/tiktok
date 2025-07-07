package data

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"user-service/internal/conf"
	"user-service/internal/data/query"
	"user-service/internal/pkg"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewUserRepo, NewDB, NewRedisClient)

// Data .
type Data struct {
	// TODO wrapped database client
	log   *log.Helper
	query *query.Query
	db    *gorm.DB
	rdb   *redis.Client
	jwt   *pkg.JWTManager
	idg   *pkg.IDGenerator
}

func NewData(db *gorm.DB, logger log.Logger, rdb *redis.Client, jwt *pkg.JWTManager, idg *pkg.IDGenerator) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 非常重要!为GEN生成的query代码设置数据库连接对象
	query.SetDefault(db)
	return &Data{query: query.Q, log: log.NewHelper(logger), rdb: rdb, jwt: jwt, idg: idg}, cleanup, nil
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
