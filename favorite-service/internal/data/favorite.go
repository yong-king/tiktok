package data

import (
	"context"
	"errors"
	pbUSer "favorite-service/api/user/v1"
	pbVideo "favorite-service/api/video/v1"
	"favorite-service/internal/data/model"
	"favorite-service/internal/data/query"
	"gorm.io/gorm"

	"favorite-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type favoriteRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewFavoriteRepo(data *Data, logger log.Logger) biz.FavoriteRepo {
	return &favoriteRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ParseToken 解析token获取用户id
func (r *favoriteRepo) ParseToken(ctx context.Context, token string, refreshToken string) (int64, error) {
	resp, err := r.data.UserClient.ParseToken(ctx, &pbUSer.ParseTokenRequest{
		Token:        token,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return 0, err
	}
	return resp.UserId, nil
}

// AddFavorite 视频点赞
func (r *favoriteRepo) AddFavorite(ctx context.Context, uid int64, vid int64) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txQuery := query.Use(tx)

		// 防止重复点赞
		isFav, err := r.IsFavorited(ctx, uid, vid)
		if err != nil {
			return err
		}
		if isFav {
			return nil
		}
		favorite := &model.Favorite{
			UserID:  uid,
			VideoID: vid,
		}

		// 点赞跟新favorite
		if err := txQuery.Favorite.WithContext(ctx).Create(favorite); err != nil {
			return err
		}

		// 更新video
		_, err = txQuery.Video.WithContext(ctx).
			Where(txQuery.Video.ID.Eq(vid)).
			UpdateSimple(txQuery.Video.FavoriteCnt.Add(1))
		if err != nil {
			return err
		}

		return nil

	})
}

// RemoveFavorite 取消点赞
func (r *favoriteRepo) RemoveFavorite(ctx context.Context, uid int64, vid int64) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txQuery := query.Use(tx)

		// 取消点赞 favorite
		if _, err := txQuery.Favorite.
			WithContext(ctx).
			Where(txQuery.Favorite.UserID.Eq(uid), txQuery.Favorite.VideoID.Eq(vid)).
			Delete(); err != nil {
			return err
		}

		// video
		if _, err := txQuery.Video.
			WithContext(ctx).
			Where(txQuery.Video.ID.Eq(vid)).
			UpdateSimple(txQuery.Video.FavoriteCnt.Sub(1)); err != nil {
			return err
		}

		return nil
	})
}

// IsFavorited 幂等，防止重复点赞
func (r *favoriteRepo) IsFavorited(ctx context.Context, uid int64, vid int64) (bool, error) {
	count, err := r.data.query.Favorite.
		WithContext(ctx).
		Where(
			r.data.query.Favorite.UserID.Eq(uid),
			r.data.query.Favorite.VideoID.Eq(vid),
		).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserFavoriteVideoIDs 根据用户id获取用户点赞视频id列表
func (r *favoriteRepo) GetUserFavoriteVideoIDs(ctx context.Context, uid int64) ([]int64, error) {
	var ids []int64

	favorites, err := r.data.query.Favorite.
		WithContext(ctx).
		Where(r.data.query.Favorite.UserID.Eq(uid)).
		Find()
	if err != nil {
		return nil, err
	}

	for _, v := range favorites {
		ids = append(ids, v.VideoID)
	}

	return ids, nil
}

// CheckUserExists 用户是否存在
func (r *favoriteRepo) CheckUserExists(ctx context.Context, uid int64) (bool, error) {
	resp, err := r.data.UserClient.CheckUserExistByUserID(ctx, &pbUSer.CheckUserExistByUserIDRequest{
		UserId: uid,
	})
	if err != nil {
		return false, err
	}
	return resp.Exist, nil
}

// BatchGetVideoInfo 批量获取视频信息
func (r *favoriteRepo) BatchGetVideoInfo(ctx context.Context, ids []int64, page int, pageSize int) ([]*pbVideo.Video, error) {
	if r.data.VideoClient == nil {
		return nil, errors.New("VIDEO_CLIENT_UNAVAILABLE, Video client is not initialized")
	}

	resp, err := r.data.VideoClient.BatchGetVideoInfo(ctx, &pbVideo.BatchGetVideoInfoRequest{
		Ids:      ids,
		Page:     int32(page),
		PageSize: int32(pageSize),
	})

	if err != nil {
		return nil, err
	}

	if resp == nil || resp.Videos == nil {
		return []*pbVideo.Video{}, nil
	}

	//r.log.WithContext(ctx).Info("-----------------len(resp.Videos)-----------------", len(resp.Videos))
	return resp.Videos, nil
}
