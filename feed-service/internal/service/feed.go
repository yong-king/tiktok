package service

import (
	"context"
	"feed-service/internal/pkg/constants"

	v1 "feed-service/api/feed/v1"
	"feed-service/internal/biz"
)

// FeedService is a feed service.
type FeedService struct {
	v1.UnimplementedFeedServiceServer

	uc *biz.FeedUsecase
}

// NewFeedService new a feed service.
func NewFeedService(uc *biz.FeedUsecase) *FeedService {
	return &FeedService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *FeedService) GetFeed(ctx context.Context, in *v1.FeedRequest) (*v1.FeedReply, error) {
	// 默认为游客模式
	var userID int64 = 0

	// 解析token
	if in.Token != "" {
		// 验证token是否合法, 调用user微服务
		uid, err := s.uc.ParesToken(ctx, in.Token, in.RefreshToken)
		if err != nil {
			return nil, err
		}
		userID = uid
	}

	//latestTime := in.LastTime
	//if latestTime == 0 {
	//	latestTime = time.Now().Unix()
	//}

	offset := in.Offset
	limit := constants.FeedPageLimit

	// videos, nextTime, err := s.uc.GetFeed(ctx, userID, latestTime)
	videos, err := s.uc.GetFeed(ctx, userID, offset, int64(limit))
	if err != nil {
		return nil, err
	}

	return &v1.FeedReply{
		Videos: videos,
	}, nil
}
