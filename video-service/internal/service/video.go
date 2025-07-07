package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/errors"
	"io"
	"net/http"

	v1 "video-service/api/video/v1"

	"video-service/internal/biz"
	params "video-service/internal/biz/params"
)

// VideoService is a greeter service.
type VideoService struct {
	v1.UnimplementedVideoServiceServer

	uc *biz.VideoUsecase
}

// NewVideoService new a video service.
func NewVideoService(uc *biz.VideoUsecase) *VideoService {
	return &VideoService{uc: uc}
}

var GlobalVideoService *VideoService

func BindVideoService(svc *VideoService) {
	GlobalVideoService = svc
}

// UploadVideo 上传视频
func (s *VideoService) UploadVideo(ctx context.Context, in *v1.UploadVideoRequest) (*v1.UploadVideoReply, error) {
	// 参数校验
	// 1. 解析token是否有效
	userID, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if in.Data == nil || in.Filename == "" || len(in.Data) == 0 {
		return nil, errors.BadRequest("UploadVideo", "视频数据或文件名不能为空")
	}

	// 3. 构建对象名（加 user_id 防止重复）
	objectName := fmt.Sprintf("video/%d/%s", userID, in.Filename)

	// 4. 上传到 MinIO（通过依赖注入拿到 uploader）
	reader := bytes.NewReader(in.Data)
	playURL, err := s.uc.UploadVideo(ctx, objectName, reader, int64(len(in.Data)), "video/mp4")
	if err != nil {
		return nil, errors.InternalServer("UPLOAD_FAIL", err.Error())
	}

	// 5. 返回
	return &v1.UploadVideoReply{PlayUrl: playURL, CoverUrl: "", Duration: 0, Message: "success"}, nil
}

// UploadVideoGin 上传视频基于gin
func (s *VideoService) UploadVideoGin(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件获取失败"})
		return
	}

	token := c.PostForm("token")
	refreshToken := c.PostForm("refreshToken")

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件打开失败"})
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件读取失败"})
		return
	}

	reply, err := s.UploadVideo(c, &v1.UploadVideoRequest{
		Data:         data,
		Filename:     file.Filename,
		Token:        token,
		RefreshToken: refreshToken,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"play_url":  reply.PlayUrl,
		"cover_url": reply.CoverUrl,
		"duration":  reply.Duration,
		"message":   reply.Message,
	})
}

// CreateVideo 创建视频，即视频信息
func (s *VideoService) CreateVideo(ctx context.Context, in *v1.CreateVideoRequest) (*v1.CreateVideoReply, error) {
	// 1. 参数校验
	fmt.Printf("CreateVideo: %v\n", in)
	userID, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	if in.Title == "" || in.PlayUrl == "" {
		return nil, errors.BadRequest("CreateVideo", "invalid params")
	}
	p := params.CreateVideoReq{
		Title:       in.Title,
		Description: in.Description,
		PlayUrl:     in.PlayUrl,
		CoverUrl:    in.CoverUrl,
		Duration:    in.Duration,
		Tags:        in.Tags,
		IsPublic:    in.IsPublic,
		IsOriginal:  in.IsOriginal,
		SourceUrl:   in.SourceUrl,
		UserID:      userID,
	}
	// 2. 创建视频
	videoID, err := s.uc.CreateVideo(ctx, p)
	if err != nil {
		return nil, err
	}
	return &v1.CreateVideoReply{VideoId: videoID}, nil
}

// ListUserVideos 获取用户的视频列表
func (s *VideoService) ListUserVideos(ctx context.Context, in *v1.ListUserVideosRequest) (*v1.ListUserVideosReply, error) {
	// 1. 参数校验
	fmt.Printf("toekn: %v\n", in.Token)
	userID, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 分页默认值处理
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 || in.PageSize > 10 {
		in.PageSize = 10
	}

	p := params.ListUserVideosRequest{
		FUserId:  in.UserId,
		Page:     in.Page,
		PageSize: in.PageSize,
		UserId:   userID,
	}

	// 2. 查阅用户视频
	videoInfo, err := s.uc.ListUserVideos(ctx, p)
	if err != nil {
		return nil, err
	}

	// 视频信息
	videos := make([]*v1.Video, 0, len(videoInfo.Videos))
	for _, v := range videoInfo.Videos {
		videos = append(videos, &v1.Video{
			Id:          v.Id,
			UserId:      v.UserId,
			PlayUrl:     v.PlayUrl,
			CoverUrl:    v.CoverUrl,
			Title:       v.Title,
			Description: v.Description,
			Duration:    v.Duration,
			Tags:        v.Tags,
			FavoriteCnt: v.FavoriteCnt,
			CommentCnt:  v.CommentCnt,
			ShareCnt:    v.ShareCnt,
			CollectCnt:  v.CollectCnt,
		})
	}

	return &v1.ListUserVideosReply{
		Videos:      videos,
		Total:       videoInfo.Total,
		CurrentPage: videoInfo.CurrentPage,
		PageSize:    videoInfo.CurrentPage,
	}, nil
}

// BatchGetVideoInfo 批量获取视频信息
func (s *VideoService) BatchGetVideoInfo(ctx context.Context, in *v1.BatchGetVideoInfoRequest) (*v1.BatchGetVideoInfoReply, error) {
	//fmt.Printf("-----------------BatchGetVideoInfo ids: %v------------------------\n", in.Ids)
	if len(in.Ids) == 0 {
		return nil, errors.BadRequest("BatchGetVideoInfo", "invalid params")
	}

	videos, err := s.uc.BatchGetVideoInfo(ctx, in.Ids, int64(in.Page), int64(in.PageSize))
	if err != nil {
		return nil, err
	}

	//fmt.Printf("----------------BatchGetVideoInfo-----------------: %v\n", videos)

	return &v1.BatchGetVideoInfoReply{Videos: videos}, nil
}

// CheckVideoExists 检查视频是否存在
func (s *VideoService) CheckVideoExists(ctx context.Context, in *v1.CheckVideoExistsRequest) (*v1.CheckVideoExistsReply, error) {
	if in.VideoId == 0 {
		return nil, errors.BadRequest("CheckVideoExists", "invalid params")
	}
	exist, err := s.uc.CheckVideoExistsByID(ctx, in.VideoId)
	if err != nil {
		return nil, err
	}
	return &v1.CheckVideoExistsReply{Exist: exist}, nil
}
