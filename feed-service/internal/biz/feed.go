package biz

import (
	"context"
	v1 "feed-service/api/feed/v1"
	pbUser "feed-service/api/user/v1"
	pbVideo "feed-service/api/video/v1"

	"github.com/go-kratos/kratos/v2/log"
)

// Greeter is a Greeter model.// GreeterRepo is a Greater repo.
type FeedRepo interface {
	GetFeedVideoList(context.Context, int64, int) ([]*v1.Video, error)
	ParesToken(context.Context, string, string) (int64, error)
	BatchGetUserInfo(context.Context, []int64) ([]*pbUser.Author, error)
	BatchGetVideoInfo(ctx context.Context, vid []int64) ([]*pbVideo.Video, error)
	BatchGetVideoCountsFromCache(context.Context, []int64) (map[int64]int64, map[int64]int64, error)
	SetVideoCountsToCache(ctx context.Context, videoID, likeCount, commentCount int64) error
	GetRecommendedVideoIDs(ctx context.Context, offset, limit int64) ([]int64, error)
	GetFeedVideoListByIDS(ctx context.Context, ids []int64) ([]*v1.Video, error)
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
func (uc *FeedUsecase) GetFeed(ctx context.Context, uid, offset, limit int64) ([]*v1.Video, error) {
	uc.log.WithContext(ctx).Infof("GetFeed: %d, %d, %d", uid, offset, limit)

	// 1. 从缓存热榜中查询
	videoIDs, err := uc.repo.GetRecommendedVideoIDs(ctx, offset, limit)
	if err != nil {
		return nil, err
	}

	// 1. 从数据库中获取信息
	videos, err := uc.repo.GetFeedVideoListByIDS(ctx, videoIDs)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetFeed: %d, %d", uid, videoIDs)
		return nil, err
	}

	if len(videos) == 0 {
		return []*v1.Video{}, nil
	}

	// 2. 作者信息
	err = uc.batchFillAuthors(ctx, videos)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetFeed: %d", uid)
		return nil, err
	}

	// 3. 点赞信息，评论信息
	err = uc.batchFillVideos(ctx, videos)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("GetFeed: %d", uid)
		return nil, err
	}

	return videos, nil
}

// batchFillVideos 批量填充视频返回信息
func (uc *FeedUsecase) batchFillVideos(ctx context.Context, videos []*v1.Video) error {

	if len(videos) == 0 {
		return nil
	}

	// 1. video id
	videoIDs := make([]int64, 0, len(videos))
	for _, v := range videos {
		videoIDs = append(videoIDs, v.VideoId)
	}

	// 2. 从缓存中直接获取评论数和点赞数
	likeCount, commentCount, err := uc.repo.BatchGetVideoCountsFromCache(ctx, videoIDs)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("BatchGetVideoCountsFromCache error: %v", err)
	}

	// 3. 填充缓冲中命中的
	missingIDs := make([]int64, 0)
	for i, vid := range videoIDs {
		v := videos[i]
		if like, ok := likeCount[vid]; ok {
			v.LikeCount = like
		} else {
			missingIDs = append(missingIDs, vid)
		}
		if comment, ok := commentCount[vid]; ok {
			v.CommentCount = comment
		}
	}

	// 缓存未命中的，兜底查 DB（或由 video-service 查 DB）

	if len(missingIDs) > 0 {
		resp, err := uc.repo.BatchGetVideoInfo(ctx, missingIDs)
		if err != nil {
			return err
		}
		// 回填缓存和视频结构
		for _, v := range resp {
			for _, vv := range videos {
				if vv.VideoId == v.Id {
					vv.CommentCount = int64(v.CommentCnt)
					vv.LikeCount = int64(v.FavoriteCnt)
				}
			}
			err := uc.repo.SetVideoCountsToCache(ctx, v.Id, int64(v.FavoriteCnt), int64(v.CommentCnt))
			if err != nil {
				uc.log.WithContext(ctx).Errorf("SetVideoCountsToCache error: %v", err)
				return err
			}
		}
	}
	return nil
}

// batchFillAuthors 批量填充视频作者信息
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

// ParesToken token解析
func (uc *FeedUsecase) ParesToken(ctx context.Context, token, refreshToken string) (int64, error) {
	uc.log.WithContext(ctx).Infof("ParesToken: %s, %s", token, refreshToken)
	uid, err := uc.repo.ParesToken(ctx, token, refreshToken)
	if err != nil {
		uc.log.WithContext(ctx).Error("ParesToken: %s, %s", token, refreshToken)
		return 0, err
	}
	return uid, nil
}
