package server

import (
	"github.com/gin-gonic/gin"
	"video-service/internal/service"
)

func NewGinServer() *gin.Engine {
	r := gin.Default()

	// 上传视频接口
	r.POST("/api/video/upload", func(c *gin.Context) {
		service.GlobalVideoService.UploadVideoGin(c)
	})

	return r
}
