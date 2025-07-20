package service

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ralation-service/internal/biz/params"

	v1 "ralation-service/api/relation/v1"
	"ralation-service/internal/biz"
)

type RelationService struct {
	v1.UnimplementedRelationServiceServer

	uc *biz.RelationUsecase
}

func NewRelationService(uc *biz.RelationUsecase) *RelationService {
	return &RelationService{uc: uc}
}

// RelationControl 用户关系操作
func (s *RelationService) RelationControl(ctx context.Context, req *v1.RelationControlRequest) (*v1.RelationControlReply, error) {
	// 1. 参数校验
	if req.ActionType != 1 && req.ActionType != 2 {
		return nil, status.Error(codes.InvalidArgument, "invalid action type")
	}

	if req.ToUserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid to_suer_id")
	}

	// 2. 解析token
	userID, err := s.uc.ParseToken(ctx, req.Token, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	if userID == req.ToUserId {
		return nil, status.Error(codes.PermissionDenied, "cannot follow yourself")
	}

	// 3. 关系操作
	err = s.uc.RelationControl(ctx, &params.RelationControl{
		ToUserId:   req.ToUserId,
		UserId:     userID,
		ActionType: req.ActionType,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RelationControlReply{Msg: "success"}, nil
}

// GetRelationListByUserID 获取用户关注列表
func (s *RelationService) GetRelationListByUserID(ctx context.Context, req *v1.GetRelationListByUserIDRequest) (*v1.GetRelationListByUserIDReply, error) {
	// 1. 参数校验
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	// 2. 解析token
	currentUserID, err := s.uc.ParseToken(ctx, req.Token, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// 3. 获取用户关注列表
	users, err := s.uc.GetRelationListByUserID(ctx, currentUserID, req.UserId)
	if err != nil {
		return nil, err
	}
	return &v1.GetRelationListByUserIDReply{User: users}, nil
}
