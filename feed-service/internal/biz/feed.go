package biz

import (
	"context"
	pbUser "feed-service/api/user/v1"
	"feed-service/internal/pkg/constants"

	v1 "feed-service/api/feed/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Greeter is a Greeter model.// GreeterRepo is a Greater repo.
type FeedRepo interface {
	GetFeedVideoList(context.Context, int64, int) ([]*v1.Video, error)
	ParesToken(context.Context, string, string) (int64, error)
	BatchGetUserInfo(context.Context, []int64) ([]*pbUser.Author, error)
}

// GreeterUsecase is a Greeter usecase.
type FeedUsecase struct {
	repo FeedRepo
	log  *log.Helper
}

// NewGreeterUsecase new a Greeter usecase.
func NewFeedUsecase(repo FeedRepo, logger log.Logger) *FeedUsecase {
	return &FeedUsecase{repo: repo, log: log.NewHelper(logger)}
}

// GetFeed 获取视频流
func (uc *FeedUsecase) GetFeed(ctx context.Context, uid, lastTime int64) ([]*v1.Video, int64, error) {
	uc.log.WithContext(ctx).Infof("GetFeed: %d, %d", uid, lastTime)
	limit := constants.FeedPageLimit // 每次返回视频数量

	// 1. 从数据库中查询
	videos, err := uc.repo.GetFeedVideoList(ctx, lastTime, limit)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetFeed: %d, %d", uid, lastTime)
		return nil, 0, err
	}

	if len(videos) == 0 {
		return []*v1.Video{}, 0, nil
	}

	// 2. 作者信息
	err = uc.batchFillAuthors(ctx, videos)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetFeed: %d, %d", uid, lastTime)
		return nil, 0, err
	}

	// 3. TODO 点赞信息，评论信息

	// 4. 下次拉取的游标
	nextTime := videos[len(videos)-1].PublishTime - 1
	return videos, nextTime, nil
}

func (uc *FeedUsecase) batchFillAuthors(ctx context.Context, videos []*v1.Video) error {
	if len(videos) == 0 {
		return nil
	}

	// 1. 收集 authorIds 去重
	authorIdSet := make(map[int64]struct{})
	for _, v := range videos {
		authorIdSet[v.AuthorId] = struct{}{}
	}

	var authorIds []int64
	for id := range authorIdSet {
		authorIds = append(authorIds, id)
	}

	// 调用user微服务批量获取用户信息
	resp, err := uc.repo.BatchGetUserInfo(ctx, authorIds)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("batchFillAuthors: %v", err)
		return err
	}

	// 3. 构建 userId -> Author 映射
	authorMap := make(map[int64]*v1.Author, len(resp))
	for _, u := range resp {
		authorMap[u.Id] = &v1.Author{
			Id:        u.Id,
			Name:      u.Name,
			AvatarUrl: u.AvatarUrl,
		}
	}

	// 4. 填充到 videos 中
	for _, v := range videos {
		if author, ok := authorMap[v.AuthorId]; ok {
			v.Author = author
		}
	}

	return nil
}

func (uc *FeedUsecase) ParesToken(ctx context.Context, token, refreshToken string) (int64, error) {
	uc.log.WithContext(ctx).Infof("ParesToken: %s, %s", token, refreshToken)
	uid, err := uc.repo.ParesToken(ctx, token, refreshToken)
	if err != nil {
		uc.log.WithContext(ctx).Error("ParesToken: %s, %s", token, refreshToken)
		return 0, err
	}
	return uid, nil
}
