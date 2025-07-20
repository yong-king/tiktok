package biz

import (
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	v1 "ralation-service/api/relation/v1"
	"ralation-service/internal/biz/params"
)

type RelationRepo interface {
	CreateRelation(ctx context.Context, userID, toUserID int64) error
	DeleteRelation(ctx context.Context, userID, toUserID int64) error
	ParseToken(ctx context.Context, token, refreshToken string) (userID int64, err error)
	CheckUserExistByUserID(ctx context.Context, toUserID int64) (bool, error)
	GetFollowList(ctx context.Context, userID, toUserID int64) (users []*params.UserInfo, err error)
}

type RelationUsecase struct {
	repo RelationRepo
	log  *log.Helper
}

// NewGreeterUsecase new a Greeter usecase.
func NewGreeterUsecase(repo RelationRepo, logger log.Logger) *RelationUsecase {
	return &RelationUsecase{repo: repo, log: log.NewHelper(logger)}
}

// RelationControl 用户关系操作
func (uc *RelationUsecase) RelationControl(ctx context.Context, param *params.RelationControl) error {
	uc.log.WithContext(ctx).Infof("User %d followed User %d", param.UserId, param.ToUserId)
	// 1. 被关在对象是否存在
	exist, err := uc.repo.CheckUserExistByUserID(ctx, param.ToUserId)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("user not exist")
	}

	switch param.ActionType {
	// 创建
	case params.Action_Type_Create:
		return uc.repo.CreateRelation(ctx, param.UserId, param.ToUserId)
	case params.Action_Type_Delete:
		return uc.repo.DeleteRelation(ctx, param.UserId, param.ToUserId)
	default:
		return errors.New("invalid action type")
	}
}

// ParseToken 解析token
func (uc *RelationUsecase) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	return uc.repo.ParseToken(ctx, token, refreshToken)
}

// GetRelationListByUserID
func (uc *RelationUsecase) GetRelationListByUserID(ctx context.Context, currentUserID, targetUserID int64) ([]*v1.User, error) {
	uc.log.WithContext(ctx).Infof("<GetRelationListByUserID> currentUserID: %d, targetUserID: %d", currentUserID, targetUserID)

	users, err := uc.repo.GetFollowList(ctx, currentUserID, targetUserID)
	if err != nil {
		return nil, err
	}

	var userList []*v1.User
	for _, u := range users {
		userList = append(userList, &v1.User{
			Id:              u.ID,
			Name:            u.Name,
			FollowCount:     int32(u.FollowCount),
			FollowerCount:   int32(u.FollowerCount),
			IsFollow:        u.IsFollow,
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			Signature:       u.Signature,
			TotalFavorited:  int32(u.TotalFavorited),
			WorkCount:       int32(u.WorkCount),
			FavoriteCount:   int32(u.FavoriteCount),
		})
	}
	return userList, nil
}
