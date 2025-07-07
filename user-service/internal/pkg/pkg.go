package pkg

import (
	"github.com/google/wire"
	"user-service/internal/conf"
)

var ProviderSet = wire.NewSet(NewJWTManagerProvider, NewIDGen)

// NewJWTManagerProvider JWT
func NewJWTManagerProvider(c *conf.JWT) *JWTManager {
	return NewJWTManager(c.Secret, c.Issuer, c.Expire)
}

// NewIDGen 雪花算法
func NewIDGen(c *conf.IDGen) *IDGenerator {
	return NewIDGenerator(c)
}
