package data

import (
	"context"
	v1 "feed-service/api/feed/v1"
	pbUser "feed-service/api/user/v1"
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

func (r *feedRepo) BatchGetUserInfo(ctx context.Context, ids []int64) ([]*pbUser.Author, error) {
	resp, err := r.data.UserClient.BatchGetUserInfo(ctx, &pbUser.BatchGetUserInfoRequest{
		AuthorIds: ids,
	})
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}
