package service

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"user-service/internal/biz"
	param "user-service/internal/biz/param"

	pb "user-service/api/user/v1"
)

type UserServiceService struct {
	pb.UnimplementedUserServiceServer
	uc *biz.UserService
}

func NewUserServiceService(uc *biz.UserService) *UserServiceService {
	return &UserServiceService{uc: uc}
}

// Register 用户注册
func (s *UserServiceService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	param := &param.RegisterParam{
		Username: req.Username,
		Password: req.Password,
	}
	reply, err := s.uc.Register(ctx, param)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterReply{StatusCode: reply.Status_code, StatusMsg: reply.Status_msg, UserId: reply.UserID, Token: reply.Token}, nil
}

// Login 用户登录
func (s *UserServiceService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	param := &param.LoginParam{
		Username: req.Username,
		Password: req.Password,
	}
	reply, err := s.uc.Login(ctx, param)
	if err != nil {
		return nil, err
	}
	return &pb.LoginReply{StatusCode: reply.Status_code, StatusMsg: reply.Status_msg, UserId: reply.UserID, Token: reply.Token}, nil
}

// UserInfo 获取用户信息
func (s *UserServiceService) UserInfo(ctx context.Context, req *pb.UserInfoRequest) (*pb.UserInfoReply, error) {
	//fmt.Printf("---------------------req-------------------: %#v \n", req)
	reply, err := s.uc.GetUserInfo(ctx, req.UserId, req.CurrentUserId)
	if err != nil {
		return nil, err
	}
	user := reply.User
	return &pb.UserInfoReply{
		StatusCode: reply.Status_code,
		StatusMsg:  reply.Status_msg,
		User: &pb.User{
			Id:              user.ID,
			Name:            user.Name,
			FollowCount:     user.FollowCount,
			FollowerCount:   user.FollowerCount,
			IsFollow:        user.IsFollow,
			Avatar:          user.Avatar,
			BackgroundImage: user.BackgroundImage,
			Signature:       user.Signature,
			TotalFavorited:  user.TotalFavorited,
			WorkCount:       user.WorkCount,
			FavoriteCount:   user.FavoriteCount,
		},
	}, nil
}

// RefreshToken 刷新token
func (s *UserServiceService) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshReply, error) {
	token, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &pb.RefreshReply{StatusCode: 200, StatusMsg: "success", Token: token}, nil
}

// ParseToken 解析token
func (s *UserServiceService) ParseToken(ctx context.Context, in *pb.ParseTokenRequest) (*pb.ParseTokenReply, error) {
	user_id, err := s.uc.ParseToken(ctx, in.Token, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &pb.ParseTokenReply{UserId: user_id}, nil
}

func (s *UserServiceService) CheckUserExistByUserID(ctx context.Context, in *pb.CheckUserExistByUserIDRequest) (*pb.CheckUserExistByUserIDReply, error) {
	fmt.Printf("CheckUserExistByUserID %v\n", in.UserId)
	exist, err := s.uc.CheckUserExistByUserID(ctx, in.UserId)
	if err != nil {
		return nil, errors.InternalServer("CHECK_USER_EXIST_FAILED", err.Error())
	}
	return &pb.CheckUserExistByUserIDReply{Exist: exist}, nil
}

// BatchGetUserInfo 批量获取用户信息
func (s *UserServiceService) BatchGetUserInfo(ctx context.Context, in *pb.BatchGetUserInfoRequest) (*pb.BatchGetUserInfoReply, error) {
	if len(in.AuthorIds) == 0 {
		return &pb.BatchGetUserInfoReply{}, nil
	}

	users, err := s.uc.BatchGetUserInfo(ctx, in.AuthorIds)
	if err != nil {
		return nil, err
	}

	var pbUsers []*pb.Author
	for _, u := range users {
		pbUsers = append(pbUsers, &pb.Author{
			Id:        u.ID,
			Name:      u.Name,
			AvatarUrl: u.AvatarURL,
		})
	}
	return &pb.BatchGetUserInfoReply{Users: pbUsers}, nil
}

// 批量获取用户详细信息
func (s *UserServiceService) BatchGetUserDetailInfo(ctx context.Context, req *pb.BatchGetUserDetailInfoRequest) (*pb.BatchGetUserDetailInfoReply, error) {
	users, err := s.uc.BatchGetUserDetailInfo(ctx, req.Ids)
	if err != nil {
		return nil, err
	}
	return &pb.BatchGetUserDetailInfoReply{User: users}, nil
}

// 更新用户信息
func (s *UserServiceService) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileReply, error) {
	// 1. token解析，参数解析
	uid, err := s.uc.ParseToken(ctx, req.Token, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	err = s.uc.UpdateUserProfile(ctx, &param.UpdateUserRequsetParam{
		ID:              uid,
		Name:            req.User.Name,
		Avatar:          req.User.Avatar,
		BackgroundImage: req.User.BackgroundImage,
		Signature:       req.User.Signature,
	})

	if err != nil {
		return nil, err
	}
	return &pb.UpdateUserProfileReply{Msg: "success"}, nil
}
