package biz

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	pb "user-service/api/user/v1"
	"user-service/internal/biz/param"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// UserService 用户相关业务逻辑封装
type UserRepo interface {
	CheckUserExist(context.Context, string) (bool, error)
	CreateUser(ctx context.Context, in *param.RegisterParam) (*param.RegisterReplyParam, error)
	GetUserByUsername(context.Context, string) (*param.UserValidateParam, error)
	GenerateTokens(context.Context, int64) (string, string, error)
	RefreshToken(context.Context, string) (string, error)
	GetUserByUserID(context.Context, int64) (*param.UserInfoParam, error)
	ParseToken(context.Context, string, string) (int64, error)
	CheckUserExistByUserID(context.Context, int64) (bool, error)
	BatchGetUserInfo(ctx context.Context, userIds []int64) ([]*param.Author, error)
	BatchGetUserDetailInfo(ctx context.Context, userIds []int64) ([]*param.UserInfoParam, error)
	UpdateUserProfile(ctx context.Context, requsetParam *param.UpdateUserRequsetParam) error
}

// UserService 用户相关业务逻辑封装
type UserService struct {
	repo UserRepo
	log  *log.Helper
}

func NewUserService(repo UserRepo, logger log.Logger) *UserService {
	return &UserService{repo: repo, log: log.NewHelper(logger)}
}

// Register 用户注册逻辑，包含用户名查重、密码加密、写入数据库等流程
func (uc *UserService) Register(ctx context.Context, g *param.RegisterParam) (*param.RegisterReplyParam, error) {
	exist, _ := uc.repo.CheckUserExist(ctx, g.Username)
	if exist {
		return nil, errors.New(400, "USER_ALREADY_EXISTS", "用户已存在")
	}

	// 密码加密
	hashed, err := hashPassword(g.Password)
	if err != nil {
		return nil, errors.New(500, "HASH_ERROR", "密码加密失败")
	}
	g.Password = hashed

	// 创建用户
	rep, err := uc.repo.CreateUser(ctx, g)
	if err != nil {
		return nil, err
	}

	return rep, nil
}

// 注册加密
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Login 用户登录
func (uc *UserService) Login(ctx context.Context, g *param.LoginParam) (*param.LoginReplyParam, error) {
	// 根据用户名称获取用户信息
	user, err := uc.repo.GetUserByUsername(ctx, g.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(401, "USER_NOT_EXISTS", "<用户不存在>")
		}
		uc.log.WithContext(ctx).Errorf("Login failed: %v", err)
		return nil, err
	}

	// 密码校验
	if !checkPassword(user.Password, g.Password) {
		return nil, errors.New(401, "LOGIN_PASSWORD_ERROR", "<密码错误>")
	}

	// 生成token
	accessToken, refreshToken, err := uc.repo.GenerateTokens(ctx, user.UserID)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("生成 token failed: %v", err)
		return nil, errors.New(500, "TOKEN_GENERATION_ERROR", "Token生成失败")
	}

	uc.log.WithContext(ctx).Infof("User login success: %d", user.UserID)
	return &param.LoginReplyParam{
		Status_code:  200,
		Status_msg:   "登录成功",
		UserID:       user.UserID,
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// 登录密码验证
func checkPassword(hashedPassword, inputPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(inputPassword))
	return err == nil
}

// RefreshToken 刷新token
func (uc *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	accessToken, err := uc.repo.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", errors.New(401, "INVALID_REFRESH_TOKEN", "重新登录")
	}
	return accessToken, nil
}

// GetUserInfo 根据查询的用户id获取用户信息
func (uc *UserService) GetUserInfo(ctx context.Context, userID, currentUserID int64) (*param.UserInfoReplyParam, error) {
	uc.log.WithContext(ctx).Debugf("GetUserInfo: userID:%d currentUserID:%d", userID, currentUserID)
	// 根据查询的用户id获取用户信息
	user, err := uc.repo.GetUserByUserID(ctx, userID)
	uc.log.WithContext(ctx).Debugf("GetUserInfo: %v, err： %v", user, err)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(401, "USER_NOT_EXISTS", "<<用户不存在>>")
		}
		uc.log.WithContext(ctx).Errorf("GetUserInfo failed: %v", err)
		return nil, errors.New(401, "GetUserInfo FAILED", "<<内部错误>>")
	}

	// 当前用户是否是该用户粉丝
	isFollow := false
	if currentUserID != 0 && userID != currentUserID {
		// TODO 去follow中查询
		//isFollow, err := uc.repo.IsFollow(ctx, userID, currentUserID)
		isFollow = true
	}

	return &param.UserInfoReplyParam{
		Status_code: 200,
		Status_msg:  "success",
		User: &param.UserInfoParam{
			ID:              user.ID,
			Name:            user.Name,
			FollowerCount:   user.FollowerCount,
			FollowCount:     user.FollowCount,
			IsFollow:        isFollow,
			Avatar:          user.Avatar,
			BackgroundImage: user.BackgroundImage,
			Signature:       user.Signature,
			TotalFavorited:  user.TotalFavorited,
			WorkCount:       user.WorkCount,
			FavoriteCount:   user.FavoriteCount,
		},
	}, nil
}

// ParseToken 解析token
func (uc *UserService) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	uc.log.WithContext(ctx).Info("ParseToken: %v", token)
	user_id, err := uc.repo.ParseToken(ctx, token, refreshToken)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("ParseToken failed: %v", err)
		return 0, err
	}
	return user_id, nil
}

// 查询用户是否存在
func (uc *UserService) CheckUserExistByUserID(ctx context.Context, userID int64) (bool, error) {
	uc.log.WithContext(ctx).Info("CheckUserExistByUserID: %v", userID)
	return uc.repo.CheckUserExistByUserID(ctx, userID)
}

func (uc *UserService) BatchGetUserInfo(ctx context.Context, userIds []int64) ([]*param.Author, error) {
	return uc.repo.BatchGetUserInfo(ctx, userIds)
}

func (uc *UserService) BatchGetUserDetailInfo(ctx context.Context, userIds []int64) ([]*pb.User, error) {
	uc.log.WithContext(ctx).Info("BatchGetUserDetailInfo: %v", userIds)
	users, err := uc.repo.BatchGetUserDetailInfo(ctx, userIds)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("BatchGetUserDetailInfo failed: %v", err)
		return nil, err
	}

	var userList []*pb.User
	for _, u := range users {
		userList = append(userList, &pb.User{
			Id:              u.ID,
			Name:            u.Name,
			FollowCount:     u.FollowCount,
			FollowerCount:   u.FollowerCount,
			IsFollow:        u.IsFollow,
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			Signature:       u.Signature,
			TotalFavorited:  u.TotalFavorited,
			WorkCount:       u.WorkCount,
			FavoriteCount:   u.FavoriteCount,
		})
	}
	return userList, nil
}

func (uc *UserService) UpdateUserProfile(ctx context.Context, param *param.UpdateUserRequsetParam) error {
	uc.log.WithContext(ctx).Debugf("UpdateUserProfile: %v", param)
	return uc.repo.UpdateUserProfile(ctx, param)
}
