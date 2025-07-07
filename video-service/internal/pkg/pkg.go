package pkg

import (
	"github.com/google/wire"
	"video-service/internal/conf"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewMinioUploaderProvider, NewJWTManagerProvider, NewIDGenerator)

// NewJWTManagerProvider JWT
func NewJWTManagerProvider(c *conf.JWT) *JWTManager {
	return NewJWTManager(c.Secret, c.Issuer, c.Expire)
}

func NewMinioUploaderProvider(c *conf.Data_MinIO) *MinioUploader {
	minio, err := NewMinioUploader(c)
	if err != nil {
		panic(err)
	}
	return minio
}
