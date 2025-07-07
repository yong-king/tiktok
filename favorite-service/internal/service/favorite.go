package service

import (
	"context"
	v1 "favorite-service/api/favorite/v1"
	"favorite-service/internal/biz"
	"github.com/go-kratos/kratos/v2/errors"
)

type FavoriteService struct {
	v1.UnimplementedFavoriteServiceServer

	uc *biz.FavoriteUsecase
}

func NewFavoriteService(uc *biz.FavoriteUsecase) *FavoriteService {
	return &FavoriteService{uc: uc}
}

// FavoriteAction 视频点赞
func (s *FavoriteService) FavoriteAction(ctx context.Context, in *v1.FavoriteActionRequest) (*v1.FavoriteActionReply, error) {
	// 1. 参数校验
	// 1.1 参数
	if in.VideoId == 0 {
		return nil, errors.New(500, "INVALID_PARAM", "video_id is required")
	}
	if in.ActionType != 1 && in.ActionType != 2 {
		return nil, errors.New(500, "INVALID_PARAM", "invalid action type")
	}
	if in.Token == "" {
		return nil, errors.New(500, "INVALID_PARAM", "请先登录！")
	}

	// 1.2 解析token
	userId, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 2. 视频点赞
	err = s.uc.FavoriteAction(ctx, userId, in.ActionType, in.VideoId)
	if err != nil {
		return nil, err
	}

	// 3. 返回响应
	return &v1.FavoriteActionReply{
		Message: "success",
	}, nil
}

// 根据用户id获取用户点赞列表
func (s *FavoriteService) GetUserFavoriteVideoList(ctx context.Context, in *v1.GetUserFavoriteVideoListRequest) (*v1.GetUserFavoriteVideoListReply, error) {
	// 1. 参数解析
	// 1.1 被查询用户的id
	if in.TargetUserId == 0 {
		return nil, errors.New(500, "INVALID_PARAM", "target_user_id is required")
	}

	// 1.2 token解析获取
	_, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 1.3 分页
	page := 1
	pageSize := 10
	if in.Page > 0 {
		page = int(in.Page)
	}
	if in.Limit < 10 && in.Limit > 0 {
		pageSize = int(in.Limit)
	}

	// 2. 基于被查询用户id获取视频信息列表
	videoList, err := s.uc.GetUserFavoriteVideoList(ctx, in.TargetUserId, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &v1.GetUserFavoriteVideoListReply{Videos: videoList}, nil
}
