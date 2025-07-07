package data

import (
	"context"
	"encoding/json"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"user-service/internal/biz"
	"user-service/internal/biz/param"
	"user-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	data *Data
	log  *log.Helper
}

// NewUserRepo .
func NewUserRepo(data *Data, logger log.Logger) biz.UserRepo {
	return &userRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// 查询用户是否存在
func (r *userRepo) CheckUserExist(ctx context.Context, userName string) (bool, error) {
	_, err := r.data.query.User.
		WithContext(ctx).
		Where(r.data.query.User.Username.Eq(userName)).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debugf("user %s not exist", userName)
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// User 是对 gen 生成的 model.User 的扩展封装
type User struct {
	model.User
	Tags  datatypes.JSON `gorm:"-" json:"tags"`  // 忽略 GORM 映射，用作逻辑字段
	Extra datatypes.JSON `gorm:"-" json:"extra"` // 同上
}

func MustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// CreateUser 创建用户
func (r *userRepo) CreateUser(ctx context.Context, in *param.RegisterParam) (*param.RegisterReplyParam, error) {
	user := &model.User{
		ID:           r.data.idg.Generate(), // 雪花算法生成用户id
		Username:     in.Username,
		PasswordHash: in.Password,
		Tags:         MustJSON([]string{}),
		Extra:        MustJSON(map[string]any{}),
	}
	err := r.data.query.User.
		WithContext(ctx).
		Create(user)
	if err != nil {
		r.log.Errorf("CreateUser err: %v", err)
		return nil, err
	}

	// 生成token
	token, _, err := r.GenerateTokens(ctx, user.ID)
	if err != nil {
		r.log.Errorf("CreateUser err: %v", err)
		return nil, err
	}

	return &param.RegisterReplyParam{Status_code: 1, Status_msg: "success", UserID: user.ID, Token: token}, nil
}

// GetUserByUsername 根据用户名称获取用户信息
func (r *userRepo) GetUserByUsername(ctx context.Context, userName string) (*param.UserValidateParam, error) {
	// 获取用户信息
	user, err := r.data.query.User.
		WithContext(ctx).
		Where(r.data.query.User.Username.Eq(userName)).
		First()
	if err != nil {
		return nil, err
	}
	return &param.UserValidateParam{UserID: user.ID, Password: user.PasswordHash}, nil
}

// GenerateTokens 生成token
func (r *userRepo) GenerateTokens(ctx context.Context, userID int64) (string, string, error) {
	return r.data.jwt.CreateToken(ctx, userID)
}

// RefreshToken 刷新token
func (r *userRepo) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	return r.data.jwt.RefreshAccessToken(ctx, refreshToken)
}

// GetUserByUserID 根据用户id获取用户信息
func (r *userRepo) GetUserByUserID(ctx context.Context, userID int64) (*param.UserInfoParam, error) {
	user, err := r.data.query.User.
		WithContext(ctx).
		Where(r.data.query.User.ID.Eq(userID)).
		First()
	r.log.Debugf("user: %v", user)
	if err != nil {
		return nil, err
	}
	// TODO
	// 查询follower表中当前用户是否是被查用户的粉丝
	return &param.UserInfoParam{ID: user.ID, Name: user.Username, FollowCount: user.FollowCount, FollowerCount: user.FollowerCount,
		Avatar: user.Avatar, BackgroundImage: user.BackgroundImage, Signature: user.Signature, TotalFavorited: user.TotalFavorited,
		WorkCount: user.WorkCount, FavoriteCount: user.FavoriteCount,
	}, nil
}

func (r *userRepo) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	t, err := r.data.jwt.VerifyAndRefreshTokens(ctx, token, refreshToken)
	if err != nil {
		return 0, err
	}
	parseToken, err := r.data.jwt.ParseToken(ctx, t)
	if err != nil {
		return 0, err
	}
	return parseToken.UserID, nil
}

func (r *userRepo) CheckUserExistByUserID(ctx context.Context, userID int64) (bool, error) {
	_, err := r.data.query.User.WithContext(ctx).Where(r.data.query.User.ID.Eq(userID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Debugf("user %s not exist", userID)
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *userRepo) BatchGetUserInfo(ctx context.Context, userIds []int64) ([]*param.Author, error) {

	users, err := r.data.query.User.
		WithContext(ctx).
		Where(r.data.query.User.ID.In(userIds...)).
		Find()
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []*param.Author{}, nil
	}

	userInfos := make([]*param.Author, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, &param.Author{
			ID:        u.ID,
			Name:      u.Username,
			AvatarURL: u.Avatar,
		})
	}

	return userInfos, nil

}

func (r *userRepo) BatchGetUserDetailInfo(ctx context.Context, userIds []int64) ([]*param.UserInfoParam, error) {
	quertQ := r.data.query.User
	users, err := quertQ.
		WithContext(ctx).
		Where(quertQ.ID.In(userIds...)).
		Find()
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []*param.UserInfoParam{}, nil
	}

	userInfos := make([]*param.UserInfoParam, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, &param.UserInfoParam{
			ID:              u.ID,
			Name:            u.Username,
			FollowCount:     u.FollowCount,
			FollowerCount:   u.FollowerCount,
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			Signature:       u.Signature,
			TotalFavorited:  u.TotalFavorited,
			WorkCount:       u.WorkCount,
			FavoriteCount:   u.FavoriteCount,
		})
	}
	return userInfos, nil
}
