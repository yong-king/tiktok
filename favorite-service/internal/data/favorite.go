package data

import (
	"context"
	"errors"
	pbUSer "favorite-service/api/user/v1"
	pbVideo "favorite-service/api/video/v1"
	"favorite-service/internal/data/model"
	"favorite-service/internal/data/query"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"

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

	// 1. 幂等，防止重复点赞
	keyUserFavorite := fmt.Sprintf("favorite:user:%d", uid)
	isMember, err := r.checkUserHaveFavorite(ctx, keyUserFavorite, vid)
	if err != nil {
		r.log.WithContext(ctx).Errorf("Redis SIsMember error for uid=%d, vid=%d: %v", uid, vid, err)
		// 出错时仍继续走 DB 检查，保证正确性
	}
	if isMember {
		return nil
	}

	// 数据库事务
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	if err != nil {
		return err
	}

	// 写入redis
	// 将 vid 加入用户点赞集合
	if err := r.data.rdb.SAdd(ctx, keyUserFavorite, vid).Err(); err != nil {
		r.log.WithContext(ctx).Errorf("Redis SAdd error for uid=%d, vid=%d: %v", uid, vid, err)
		return err
	}

	// 视频点赞数自增
	keyVideoFavorite := fmt.Sprintf("video:favorite:%d", vid)

	// 检查是否存在，若不存在则从 DB 回填
	err = r.checkVidoeFavoriteInCache(ctx, keyVideoFavorite, vid)
	if err != nil {
		r.log.WithContext(ctx).Errorf("checkVideoFavoriteInCache error for vid=%d: %v", vid, err)
		return err
	}
	if err := r.data.rdb.Incr(ctx, keyVideoFavorite).Err(); err != nil {
		r.log.WithContext(ctx).Errorf("Redis Incr error for vid=%d: %v", vid, err)
		return err
	}

	err = r.UpdateVideoScoreAfterLike(ctx, vid)
	if err != nil {
		return err
	}

	return nil
}

// RemoveFavorite 取消点赞
func (r *favoriteRepo) RemoveFavorite(ctx context.Context, uid int64, vid int64) error {

	// 数据库事务操作
	err := r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txQuery := query.Use(tx)

		// 取消点赞 favorite
		result, err := txQuery.Favorite.
			WithContext(ctx).
			Where(txQuery.Favorite.UserID.Eq(uid), txQuery.Favorite.VideoID.Eq(vid)).
			Delete()
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return nil // 幂等
		}

		// video
		result, err = txQuery.Video.
			WithContext(ctx).
			Where(txQuery.Video.ID.Eq(vid)).
			UpdateSimple(txQuery.Video.FavoriteCnt.Sub(1))
		if err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return errors.New("video not found or favorite count is 0")
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 同步更新redis
	keyUserFavorite := fmt.Sprintf("favorite:user:%d", uid)
	if err := r.data.rdb.SRem(ctx, keyUserFavorite, vid).Err(); err != nil {
		r.log.WithContext(ctx).Errorf("Redis SRem error for uid=%d, vid=%d: %v", uid, vid, err)
		return err
	}

	keyVideoFavorite := fmt.Sprintf("video:favorite:%d", vid)
	err = r.checkVidoeFavoriteInCache(ctx, keyVideoFavorite, vid)
	if err != nil {
		r.log.WithContext(ctx).Errorf("Redis SIsMember error: %v", err)
		return err
	}
	if err := r.data.rdb.Decr(ctx, keyVideoFavorite).Err(); err != nil {
		r.log.WithContext(ctx).Errorf("Redis Decr error: %v", err)
		return err
	}

	err = r.UpdateVideoScoreAfterLike(ctx, vid)
	if err != nil {
		return err
	}

	return nil
}

// UpdateVideoScoreAfterLike 更新分数
func (r *favoriteRepo) UpdateVideoScoreAfterLike(ctx context.Context, videoID int64) error {
	video, err := r.data.VideoClient.GetVideoFavoriteAndCommentCount(ctx, &pbVideo.GetVideoFavoriteAndCommentCountRequest{VideoId: videoID}) // 获取点赞数、评论数、上传时间
	if err != nil {
		r.log.Errorf("GetVideo error: %v", err)
		return err
	}
	score, err := r.data.VideoClient.CalcVideoScore(ctx, &pbVideo.CalcVideoScoreRequest{
		FavoriteCount: video.FavoriteCount,
		CommentCount:  video.CommentCount,
		UploadTime:    video.UploadTime,
	})
	if err != nil {
		r.log.Errorf("CalcVideoScore error: %v", err)
		return err
	}
	err = r.data.rdb.ZAdd(ctx, "video:score", redis.Z{
		Score:  float64(score.Score),
		Member: videoID,
	}).Err()
	if err != nil {
		r.log.Errorf("ZAdd video:score error: %v", err)
		return err
	}
	return nil
}

// checkUserHaveFavorite 是否已经点赞
func (r *favoriteRepo) checkUserHaveFavorite(ctx context.Context, keyUserFavorite string, vid int64) (bool, error) {
	return r.data.rdb.SIsMember(ctx, keyUserFavorite, vid).Result()
}

// 是否有点赞在缓存中
func (r *favoriteRepo) checkVidoeFavoriteInCache(ctx context.Context, keyVideoFavorite string, vid int64) error {
	val, err := r.data.rdb.Exists(ctx, keyVideoFavorite).Result()
	if err != nil {
		r.log.WithContext(ctx).Errorf("Redis Exists error: %v", err)
		return err
	}
	if val == 0 {
		// Redis 不存在，回源 DB
		var cnt int64
		err = r.data.query.WithContext(ctx).Video.Where(query.Video.ID.Eq(vid)).Select(query.Video.FavoriteCnt).Scan(&cnt)
		if err != nil {
			return err
		}
		err = r.data.rdb.Set(ctx, keyVideoFavorite, cnt, 24*time.Hour).Err()
		if err != nil {
			return err
		}
	}
	return nil
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
