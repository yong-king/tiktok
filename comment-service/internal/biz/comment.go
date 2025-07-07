package biz

import (
	"comment-service/internal/biz/param"
	"comment-service/internal/data/model"
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
)

type CommentRepo interface {
	ParseToken(ctx context.Context, token, refreshToken string) (int64, error)
	CreateComment(ctx context.Context, req *param.CreateCommentRequest) (*param.CreateCommentResponse, error)
	DeleteComment(ctx context.Context, commentId int64, videoId int64) error
	CheckVideoExist(ctx context.Context, videoId int64) (bool, error)
	GetCommentList(ctx context.Context, videoId, page, pageSize int64) ([]*model.Comment, error)
}

type CommentUsecase struct {
	repo CommentRepo
	log  *log.Helper
}

func NewCommentUsecase(repo CommentRepo, logger log.Logger) *CommentUsecase {
	return &CommentUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ParseToken 解析token返回uid
func (uc *CommentUsecase) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	uc.log.WithContext(ctx).Info("ParseToken", "token", token, "refresh_token", refreshToken)
	return uc.repo.ParseToken(ctx, token, refreshToken)
}

// CreateComment 创建和删除评论
func (uc *CommentUsecase) CreateComment(ctx context.Context, req *param.CreateCommentRequest) (*param.CreateCommentResponse, error) {
	uc.log.WithContext(ctx).Info("CreateComment", "req", req)

	// 检查被评论的视频是否存在
	exists, err := uc.repo.CheckVideoExist(ctx, req.VideoId)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("video not exist")
	}

	// 根据action_type对评论进行操作
	switch req.ActionType {
	case 1:
		// 为 1 为创建评论
		resp, err := uc.repo.CreateComment(ctx, req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	case 2:
		// 为 2 为删除评论
		err := uc.repo.DeleteComment(ctx, req.CommentID, req.VideoId)
		if err != nil {
			return nil, err
		}
		return &param.CreateCommentResponse{
			CommentID: req.CommentID,
			Message:   "delete comment success",
		}, nil
	default:
		return nil, errors.New("invalid action type")
	}
}

func (uc *CommentUsecase) GetCommentList(ctx context.Context, videoId, page, pageSize int64) ([]*model.Comment, error) {
	uc.log.WithContext(ctx).Info("GetCommentList:", "videoId", videoId, "pageSize", pageSize, "page", page)
	// 1. 检查视频视频存在
	exists, err := uc.repo.CheckVideoExist(ctx, videoId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("video not exist")
	}

	// 2. 获取视频评论信息
	return uc.repo.GetCommentList(ctx, videoId, page, pageSize)
}
