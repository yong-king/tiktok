package data

import (
	"context"
	v1 "feed-service/api/feed/v1"
	pbUser "feed-service/api/user/v1"
	pbVideo "feed-service/api/video/v1"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"

	"feed-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type feedRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewFeedRepo(data *Data, logger log.Logger) biz.FeedRepo {
	return &feedRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// GetFeedVideoList 获取视频流
func (r *feedRepo) GetFeedVideoList(ctx context.Context, lastTime int64, limit int) ([]*v1.Video, error) {

	queryTime := time.Unix(lastTime, 0)
	videos, err := r.data.query.Video.
		WithContext(ctx).
		Where(r.data.query.Video.CreatedAt.Lte(queryTime)).
		Order(r.data.query.Video.CreatedAt.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		return nil, err
	}

	var results []*v1.Video
	for _, video := range videos {
		results = append(results, &v1.Video{
			VideoId:      video.ID,
			Title:        video.Title,
			CoverUrl:     video.CoverURL,
			AuthorId:     video.UserID,
			LikeCount:    int64(video.FavoriteCnt),
			CommentCount: int64(video.CommentCnt),
			PublishTime:  video.CreatedAt.Unix(),
			// TODO 当前用户是否点赞
		})
	}

	return results, nil
}

// ParesToken tokne解析
func (r *feedRepo) ParesToken(ctx context.Context, token string, refreshToken string) (int64, error) {
	rep, err := r.data.UserClient.ParseToken(ctx, &pbUser.ParseTokenRequest{
		Token:        token,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return 0, err
	}
	return rep.UserId, nil
}

// BatchGetUserInfo 批量获取作者信息
func (r *feedRepo) BatchGetUserInfo(ctx context.Context, ids []int64) ([]*pbUser.Author, error) {
	resp, err := r.data.UserClient.BatchGetUserInfo(ctx, &pbUser.BatchGetUserInfoRequest{
		AuthorIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}

// BatchGetVideoInfo 根据视频id批量获取视频信息
func (r *feedRepo) BatchGetVideoInfo(ctx context.Context, ids []int64) ([]*pbVideo.Video, error) {
	resp, err := r.data.VideoClient.BatchGetVideoInfo(ctx, &pbVideo.BatchGetVideoInfoRequest{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}
	return resp.Videos, nil
}

// BatchGetVideoCountsFromCache 批量从缓存中获取点赞和批量数量信息
func (r *feedRepo) BatchGetVideoCountsFromCache(ctx context.Context, ids []int64) (map[int64]int64, map[int64]int64, error) {

	likeCounts := make(map[int64]int64)
	commentCounts := make(map[int64]int64)

	pipe := r.data.rdb.Pipeline()
	likeCmds := make([]*redis.StringCmd, len(ids))
	commentCmds := make([]*redis.StringCmd, len(ids))

	for i, id := range ids {
		likeCmds[i] = pipe.Get(ctx, fmt.Sprintf("video:favorite:%d", id))
		commentCmds[i] = pipe.Get(ctx, fmt.Sprintf("video:comment_count:%d", id))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, nil, err
	}
	for i, id := range ids {
		if val, err := likeCmds[i].Int64(); err == nil {
			likeCounts[id] = val
		}
		if val, err := commentCmds[i].Int64(); err == nil {
			commentCounts[id] = val
		}
	}
	return likeCounts, commentCounts, nil
}

// 未在缓存中的加入缓存
func (r *feedRepo) SetVideoCountsToCache(ctx context.Context, videoID, likeCount, commentCount int64) error {
	pipe := r.data.rdb.Pipeline()
	pipe.Set(ctx, fmt.Sprintf("video:like_count:%d", videoID), likeCount, 24*time.Hour)
	pipe.Set(ctx, fmt.Sprintf("video:comment_count:%d", videoID), commentCount, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// GetRecommendedVideoIDs 从缓存中获取视频id的排行
func (r *feedRepo) GetRecommendedVideoIDs(ctx context.Context, offset, limit int64) ([]int64, error) {
	idsStr, err := r.data.rdb.ZRevRange(ctx, "video:score", offset, offset+limit-1).Result()
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, idStr := range idsStr {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *feedRepo) GetFeedVideoListByIDS(ctx context.Context, ids []int64) ([]*v1.Video, error) {
	videos, err := r.data.query.Video.WithContext(ctx).Where(r.data.query.Video.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	// 按 ids 顺序排回
	idToVideo := map[int64]*v1.Video{}
	for _, video := range videos {
		idToVideo[video.ID] = &v1.Video{
			VideoId:      video.ID,
			Title:        video.Title,
			CoverUrl:     video.CoverURL,
			AuthorId:     video.UserID,
			LikeCount:    int64(video.FavoriteCnt),
			CommentCount: int64(video.CommentCnt),
			PublishTime:  video.CreatedAt.Unix(),
		}
	}

	var ordered []*v1.Video
	for _, video := range ids {
		if val, ok := idToVideo[video]; ok {
			ordered = append(ordered, val)
		}
	}

	return ordered, nil
}
