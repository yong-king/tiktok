package pkg

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"video-service/internal/conf"
)

type MinioUploader struct {
	client     *minio.Client // MinIO 客户端
	bucketName string        // 存储桶名称
	endpoint   string        // MinIO 访问地址（含端口）
}

// NewMinioUploader 初始化 MinIO 客户端并确保 bucket 存在
func NewMinioUploader(cfg *conf.Data_MinIO) (*MinioUploader, error) {
	// 创建 MinIO 客户端
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	// 检查 bucket 是否存在，不存在则创建
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &MinioUploader{
		client:     client,
		bucketName: cfg.BucketName,
		endpoint:   cfg.Endpoint,
	}, nil
}

// Upload 上传文件到 MinIO 并返回外部可访问的 URL
func (u *MinioUploader) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	// 上传对象
	_, err := u.client.PutObject(ctx, u.bucketName, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	// 构建播放地址（假设 MinIO 配置了公共访问）
	playURL := fmt.Sprintf("http://%s/%s/%s", u.endpoint, u.bucketName, objectName)
	return playURL, nil
}
