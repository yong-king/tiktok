package biz

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"io"
	pbUser "video-service/api/user/v1"
	v1 "video-service/api/video/v1"
	"video-service/internal/biz/params"
)

// GreeterRepo is a Greater repo.
type VideoRepo interface {
	ParseToken(context.Context, string) (int64, error)
	VerifyAndRefreshTokens(context.Context, string, string) (string, error)
	UploadVideo(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	CheckVideoExist(context.Context, string, int64) (bool, error)
	CreateVideo(context.Context, *params.CreateVideoReq) (int64, error)
	ListUserVideos(context.Context, int64, int32, int32) ([]*params.Video, int32, error)
	CheckUserExistByUserID(context.Context, int64) (*pbUser.CheckUserExistByUserIDReply, error)
	BatchGetVideoInfo(context.Context, []int64, int64, int64) ([]*v1.Video, error)
	CheckVideoExistsByID(ctx context.Context, videoID int64) (bool, error)
}

// VideoUsecase is a Video usecase.
type VideoUsecase struct {
	repo VideoRepo
	log  *log.Helper
}

// NewVideoUsecase new a Video usecase.
func NewVideoUsecase(repo VideoRepo, logger log.Logger) *VideoUsecase {
	return &VideoUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ParseToken 解析token
func (uc *VideoUsecase) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	uc.log.Infof("token: %v, refreshToken: %v", token, refreshToken)
	newToken, err := uc.repo.VerifyAndRefreshTokens(ctx, token, refreshToken)
	if err != nil {
		return 0, err
	}
	userID, err := uc.repo.ParseToken(ctx, newToken)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// UploadVideo 上传视频
func (uc *VideoUsecase) UploadVideo(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	playURL, err := uc.repo.UploadVideo(ctx, objectName, reader, size, contentType)
	if err != nil {
		return "", fmt.Errorf("usecase upload video failed: %w", err)
	}
	return playURL, nil
}

// CreateVideo 创建视频
func (uc *VideoUsecase) CreateVideo(ctx context.Context, params params.CreateVideoReq) (int64, error) {
	// 1. 改视频之前上传过没有
	exist, err := uc.repo.CheckVideoExist(ctx, params.PlayUrl, params.UserID)
	if err != nil {
		return 0, errors.InternalServer("QUERY_ERROR", err.Error())
	}
	if exist {
		return 0, errors.BadRequest("VIDEO_ALREADY_EXIST", "video already exists")
	}
	// TODO 敏感词汇检测过滤
	// 2. 雪花算法生成videoID
	// 3. 上传视频信息
	videoID, err := uc.repo.CreateVideo(ctx, &params)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("create video error: %v", err)
		return 0, errors.InternalServer("CREATE_VIDEO_FAILED", err.Error())
	}
	return videoID, nil
}

// ListUserVideos 根据用户id获取用户视频列表
func (uc *VideoUsecase) ListUserVideos(ctx context.Context, p params.ListUserVideosRequest) (params.ListUserVideosReply, error) {
	// 1. 参数校验
	// 1.1 被查询用户的userid是否合法
	resp, err := uc.repo.CheckUserExistByUserID(ctx, p.FUserId)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("调用 user 服务查询用户信息失败: %v", err)
		return params.ListUserVideosReply{}, errors.InternalServer("USER_SERVICE_ERROR", "查询用户服务失败")
	}
	if !resp.Exist {
		return params.ListUserVideosReply{}, errors.NotFound("USER_NOT_FOUND", "用户不存在")
	}

	// 2. 根据被查询用户的userid查找视频列表
	videos, total, err := uc.repo.ListUserVideos(ctx, p.FUserId, p.Page, p.PageSize)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("ListUserVideos repo error: %v", err)
		return params.ListUserVideosReply{}, errors.InternalServer("LIST_USER_VIDEOS_FAILED", err.Error())
	}

	return params.ListUserVideosReply{
		Videos:      videos,
		Total:       total,
		CurrentPage: p.Page,
		PageSize:    p.PageSize,
	}, nil
}

func (uc *VideoUsecase) BatchGetVideoInfo(ctx context.Context, ids []int64, page, pageSize int64) ([]*v1.Video, error) {
	uc.log.WithContext(ctx).Infof("BatchGetVideoInfo: %v", ids)
	return uc.repo.BatchGetVideoInfo(ctx, ids, page, pageSize)
}

func (uc *VideoUsecase) CheckVideoExistsByID(ctx context.Context, videoID int64) (bool, error) {
	return uc.repo.CheckVideoExistsByID(ctx, videoID)
}
