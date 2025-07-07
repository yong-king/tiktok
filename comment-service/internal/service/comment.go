package service

import (
	v1 "comment-service/api/comment/v1"
	"comment-service/internal/biz"
	"comment-service/internal/biz/param"
	"context"
	"github.com/go-kratos/kratos/v2/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommentService struct {
	v1.UnimplementedCommentServiceServer

	uc *biz.CommentUsecase
}

// NewGreeterService new a greeter service.
func NewCommentService(uc *biz.CommentUsecase) *CommentService {
	return &CommentService{uc: uc}
}

func (s *CommentService) CreateComment(ctx context.Context, in *v1.CreateCommentRequest) (*v1.CreateCommentReply, error) {
	// 1. 参数解析
	// 1.1 请求参数
	if in.ActionType != 1 && in.ActionType != 2 || in.VideoId == 0 || len(in.Token) == 0 {
		return nil, errors.New(500, "INVALID_PARAM", "ActionType VideoId Token IS NECESSARY")
	}

	// 1.2 处理token，解析出用户id
	uid, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 2. 添加评论
	resp, err := s.uc.CreateComment(ctx, &param.CreateCommentRequest{
		ActionType: in.ActionType,
		VideoId:    in.VideoId,
		UserID:     uid,
		Content:    in.Content,
		CommentID:  in.CommentId,
		ParentID:   in.ParentId,
	})
	if err != nil {
		return nil, err
	}
	// 3. 返回响应
	return &v1.CreateCommentReply{
		CommentId: resp.CommentID,
		Msg:       resp.Message,
	}, nil
}

// GetCommentList 获取视频评论列表
func (s *CommentService) GetCommentList(ctx context.Context, in *v1.GetCommentListRequest) (*v1.GetCommentListReply, error) {
	// 1. 参数校验
	if in.VideoId == 0 {
		return nil, errors.New(500, "INVALID_PARAM", "VideoId is zero")
	}

	_, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}

	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	comments, err := s.uc.GetCommentList(ctx, in.VideoId, page, pageSize)
	if err != nil {
		return nil, err
	}
	res := make([]*v1.Comment, 0, len(comments))
	for _, comment := range comments {
		res = append(res, &v1.Comment{
			Id:        comment.ID,
			UserId:    comment.UserID,
			VideoId:   comment.VideoID,
			ParentId:  comment.ParentID,
			Content:   comment.Content,
			CreatedAt: timestamppb.New(comment.CreatedAt),
		})
	}
	return &v1.GetCommentListReply{Comments: res}, nil
}
