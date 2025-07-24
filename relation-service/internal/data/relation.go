package data

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	pbUser "ralation-service/api/user/v1"
	"ralation-service/internal/biz/params"
	"ralation-service/internal/data/model"
	"ralation-service/internal/data/query"
	"time"

	"ralation-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type relationRepo struct {
	data *Data
	log  *log.Helper
}

func NewRelationRepo(data *Data, logger log.Logger) biz.RelationRepo {
	return &relationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// CheckUserExistByUserID 检查用户id合法性
func (r *relationRepo) CheckUserExistByUserID(ctx context.Context, userId int64) (bool, error) {
	resp, err := r.data.UserClient.CheckUserExistByUserID(ctx, &pbUser.CheckUserExistByUserIDRequest{
		UserId: userId,
	})
	if err != nil {
		return false, err
	}
	return resp.Exist, nil
}

// ParseToken 解析token，获取user_id
func (r *relationRepo) ParseToken(ctx context.Context, token, refreshToken string) (userID int64, err error) {
	resp, err := r.data.UserClient.ParseToken(ctx, &pbUser.ParseTokenRequest{
		RefreshToken: refreshToken,
		Token:        token,
	})
	if err != nil {
		return 0, err
	}
	return resp.UserId, nil
}

// CreateRelation 建立关系
func (r *relationRepo) CreateRelation(ctx context.Context, userID, toUserID int64) error {
	r.log.WithContext(ctx).Infof("User %d followed user %d", userID, toUserID)

	// 1. 判断是否已经建立了关系
	exist, err := r.CheckRelationExist(ctx, userID, toUserID)
	if err != nil {
		return err
	}

	// 事务完成后，同步更新 Redis（最终一致，不影响主流程）
	go func() {
		keyFollowing := fmt.Sprintf("user:folloing:%d", userID)
		keyFollower := fmt.Sprintf("user:folloer:%d", toUserID)
		relationKey := fmt.Sprintf("relation:%d:%d", userID, toUserID)

		pipe := r.data.rdb.TxPipeline()
		pipe.SAdd(ctx, keyFollowing, toUserID)
		pipe.SAdd(ctx, keyFollower, userID)
		pipe.Set(ctx, relationKey, "1", 24*time.Hour)
		pipe.Expire(ctx, keyFollowing, 24*time.Hour)
		pipe.Expire(ctx, keyFollower, 24*time.Hour)

		_, err := pipe.Exec(ctx)
		if err != nil {
			r.log.WithContext(ctx).Errorf("pipe.SAdd: %v", err)
		}
	}()

	if exist {
		return r.data.db.Transaction(func(tx *gorm.DB) error {
			queryTx := query.Use(tx)

			// 恢复软删除关系
			_, err := queryTx.Relation.WithContext(ctx).
				Where(
					queryTx.Relation.UserID.Eq(userID),
					queryTx.Relation.ToUserID.Eq(toUserID),
					queryTx.Relation.DeletedAt.IsNotNull(),
				).
				Update(queryTx.Relation.DeletedAt, nil)
			if err != nil {
				return err
			}

			if err := r.updateFollowStats(ctx, tx, userID, toUserID, 1); err != nil {
				return err
			}

			return nil
		})
	} else {
		// 2. 建立关系
		return r.data.db.Transaction(func(tx *gorm.DB) error {
			queryTx := query.Use(tx)

			// 添加relation表
			if err := queryTx.Relation.
				WithContext(ctx).
				Create(&model.Relation{UserID: userID, ToUserID: toUserID}); err != nil {
				return err
			}

			if err := r.updateFollowStats(ctx, tx, userID, toUserID, 1); err != nil {
				return err
			}

			return nil
		})
	}

}

func (r *relationRepo) updateFollowStats(ctx context.Context, tx *gorm.DB, userID, toUserID int64, delta int32) error {
	queryTx := query.Use(tx)

	// 更新被关注用户粉丝数量
	if _, err := queryTx.User.
		WithContext(ctx).
		Where(queryTx.User.ID.Eq(toUserID)).
		UpdateSimple(queryTx.User.FollowerCount.Add(delta)); err != nil {
		return err
	}

	// 更新当前用户关注数量
	if _, err := queryTx.User.
		WithContext(ctx).
		Where(queryTx.User.ID.Eq(userID)).
		UpdateSimple(queryTx.User.FollowCount.Add(delta)); err != nil {
		return err
	}

	return nil
}

// DeleteRelation 删除关系
func (r *relationRepo) DeleteRelation(ctx context.Context, userID, toUserID int64) error {
	// 1. 判断是否已经建立了关系
	exist, err := r.CheckRelationExist(ctx, userID, toUserID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("relation not exist")
	}

	// 2. 删除关系
	err = r.data.db.Transaction(func(tx *gorm.DB) error {
		txQuery := query.Use(tx)

		// 删除用户关系
		if _, err := txQuery.Relation.
			WithContext(ctx).
			Where(
				txQuery.Relation.UserID.Eq(userID),
				txQuery.Relation.ToUserID.Eq(toUserID),
				txQuery.Relation.DeletedAt.IsNull()).
			Update(txQuery.Relation.DeletedAt, time.Now()); err != nil {
			return err
		}

		// 更新user表中的粉丝数和关注数
		if _, err := txQuery.User.
			WithContext(ctx).
			Where(txQuery.User.ID.Eq(toUserID)).
			UpdateSimple(txQuery.User.FollowerCount.Sub(1)); err != nil {
			return err
		}

		if _, err := txQuery.User.
			WithContext(ctx).
			Where(txQuery.User.ID.Eq(userID)).
			UpdateSimple(txQuery.User.FollowCount.Sub(1)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		keyFollowing := fmt.Sprintf("user:following:%d", userID)
		keyFollower := fmt.Sprintf("user:follower:%d", toUserID)
		relationKey := fmt.Sprintf("relation:%d:%d", userID, toUserID)

		pipe := r.data.rdb.TxPipeline()

		pipe.SRem(ctx, keyFollowing, toUserID) // 从 following 集合移除被关注人
		pipe.SRem(ctx, keyFollower, userID)    // 从 follower 集合移除关注人
		pipe.Set(ctx, relationKey, "0", 24*time.Hour)
		_, err := pipe.Exec(ctx)
		if err != nil {
			r.log.WithContext(ctx).Errorf("pipe.Del: %v", err)
		}
	}()

	return nil
}

// CheckRelationExist 关系是否存在
func (r *relationRepo) CheckRelationExist(ctx context.Context, userID, toUserID int64) (bool, error) {
	key := fmt.Sprintf("relation:%d:%d", userID, toUserID)

	// 1. 先查询redis
	val, err := r.data.rdb.Get(ctx, key).Result()
	if err == nil {
		// 命中
		return val == "1", nil
	}

	if err != redis.Nil {
		r.log.WithContext(ctx).Errorf("redis get error: %v", err)
		return false, nil
	}

	// 2. 缓存未中，查询es
	exists, err := r.queryRelationExistInES(ctx, userID, toUserID)
	if err != nil {
		r.log.WithContext(ctx).Errorf("queryRelationExistInES: %v", err)
		//return false, err
	}
	if exists {
		// 回填 Redis，防止缓存击穿
		_ = r.data.rdb.Set(ctx, key, "1", 10*time.Minute).Err()
		return true, nil
	}

	// 3. 到数据库查询
	_, err = r.data.query.Relation.
		WithContext(ctx).
		Where(r.data.query.Relation.UserID.Eq(userID), r.data.query.Relation.ToUserID.Eq(toUserID)).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 回填 Redis 为不存在
			_ = r.data.rdb.Set(ctx, key, "0", 10*time.Minute).Err()
			return false, nil
		}
		r.log.WithContext(ctx).Error(err)
		return false, err
	}

	// 若 DB 查询成功，回填 Redis 为存在
	_ = r.data.rdb.Set(ctx, key, "1", 10*time.Minute).Err()
	return true, nil
}

func (r *relationRepo) queryRelationExistInES(ctx context.Context, userID, toUserID int64) (bool, error) {

	resq, err := r.data.es.Search().Index(r.data.esIndex).Query(
		&types.Query{
			Bool: &types.BoolQuery{
				Must: []types.Query{
					{
						Term: map[string]types.TermQuery{
							"user_id": {Value: userID},
						},
					},
					{
						Term: map[string]types.TermQuery{
							"to_user_id": {
								Value: toUserID},
						},
					},
				},
			},
		}).Do(ctx)
	if err != nil {
		r.log.WithContext(ctx).Errorf("queryRelationExistInES es search error: %v", err)
		return false, err
	}
	return resq.Hits.Total.Value > 0, nil
}

// GetFollowList 后去粉丝列表
func (r *relationRepo) GetFollowList(ctx context.Context, currentUserID, targetUserID int64) ([]*params.UserInfo, error) {
	queryQ := r.data.query

	// 1. 查询targetUserID关注的用户id列表
	relations, err := queryQ.Relation.
		WithContext(ctx).
		Select(queryQ.Relation.ToUserID).
		Where(queryQ.Relation.UserID.Eq(targetUserID)).
		Find()
	if err != nil {
		return nil, err
	}
	if len(relations) == 0 {
		return []*params.UserInfo{}, nil
	}

	// 提取被关注的 user_id 列表
	var followIDs []int64
	for _, rel := range relations {
		followIDs = append(followIDs, rel.ToUserID)
	}

	// 2.批量查询用户详细信息
	resp, err := r.data.UserClient.BatchGetUserDetailInfo(ctx, &pbUser.BatchGetUserDetailInfoRequest{Ids: followIDs})
	if err != nil {
		return nil, err
	}
	users := resp.GetUser()

	// 3. 查询这些用户是否被currentUserID关注
	relationList, err := queryQ.Relation.
		WithContext(ctx).
		Select(queryQ.Relation.ToUserID).
		Where(queryQ.Relation.UserID.Eq(currentUserID), queryQ.Relation.ToUserID.In(followIDs...)).
		Find()
	if err != nil {
		return nil, err
	}

	followMap := make(map[int64]bool)
	for _, rel := range relationList {
		followMap[rel.ToUserID] = true
	}

	// 4. 返回
	var res []*params.UserInfo
	for _, u := range users {
		res = append(res, &params.UserInfo{
			ID:              u.Id,
			Name:            u.Name,
			FollowCount:     int64(u.FollowCount),
			FollowerCount:   int64(u.FollowerCount),
			Avatar:          u.Avatar,
			BackgroundImage: u.BackgroundImage,
			Signature:       u.Signature,
			TotalFavorited:  int64(u.TotalFavorited),
			WorkCount:       int64(u.WorkCount),
			FavoriteCount:   int64(u.FavoriteCount),
			IsFollow:        followMap[u.Id],
		})
	}
	return res, nil
}
