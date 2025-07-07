package pkg

import (
	"comment-service/internal/conf"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewIDGen)

// NewIDGen 雪花算法
func NewIDGen(c *conf.IDGen) *IDGenerator {
	return NewIDGenerator(c)
}
