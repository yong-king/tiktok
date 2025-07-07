package biz

import (
	"context"
	"errors"
	v1 "favorite-service/api/favorite/v1"
	pbVideo "favorite-service/api/video/v1"
	"github.com/go-kratos/kratos/v2/log"
)

// GreeterRepo is a Greater repo.
type FavoriteRepo interface {
	ParseToken(context.Context, string, string) (int64, error)
	AddFavorite(ctx context.Context, uid int64, vid int64) error
	RemoveFavorite(ctx context.Context, uid int64, vid int64) error
	GetUserFavoriteVideoIDs(ctx context.Context, uid int64) ([]int64, error)
	CheckUserExists(ctx context.Context, uid int64) (bool, error)
	BatchGetVideoInfo(ctx context.Context, ids []int64, page int, pageSize int) ([]*pbVideo.Video, error)
}

type FavoriteUsecase struct {
	repo FavoriteRepo
	log  *log.Helper
}

func NewFavoriteUsecase(repo FavoriteRepo, logger log.Logger) *FavoriteUsecase {
	return &FavoriteUsecase{repo: repo, log: log.NewHelper(logger)}
}

// ParseToken 解析token获取用户id
func (uc *FavoriteUsecase) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	uc.log.WithContext(ctx).Infof("<ParseToken>token%v", token)
	uid, err := uc.repo.ParseToken(ctx, token, refreshToken)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

// FavoriteAction 视频点赞操作
func (uc *FavoriteUsecase) FavoriteAction(ctx context.Context, uid int64, actionType int32, vid int64) error {
	uc.log.WithContext(ctx).Infof("FavoriteAction: user_id=%d action_type=%d video_id=%d", uid, actionType, vid)

	switch actionType {
	// action_type=1，为点赞
	case 1:
		return uc.repo.AddFavorite(ctx, uid, vid)
	// action_type=2，为取消点赞
	case 2:
		return uc.repo.RemoveFavorite(ctx, uid, vid)
	default:
		return errors.New("INVALID_PARAM")
	}

}

func (uc *FavoriteUsecase) GetUserFavoriteVideoList(ctx context.Context, uid int64, page, pageSize int) ([]*v1.Video, error) {
	uc.log.WithContext(ctx).Infof("GetUserFavoriteVideoList: uid=%d", uid)
	// 1. 检查用户是否存在
	exists, err := uc.repo.CheckUserExists(ctx, uid)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("user not exists")
	}

	// 2. 根据uid获取视频ids
	ids, err := uc.repo.GetUserFavoriteVideoIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*v1.Video{}, nil
	}

	// 3. 根据获取的视频ids批量查询视频信息（video-service）
	uc.log.WithContext(ctx).Infof("GetUserFavoriteVideoList: ids=%v", ids)
	videos, err := uc.repo.BatchGetVideoInfo(ctx, ids, page, pageSize)
	if err != nil {
		return nil, err
	}
	videoList := make([]*v1.Video, 0, len(videos))
	for _, v := range videos {
		videoList = append(videoList, &v1.Video{
			VideoId:      v.Id,
			Title:        v.Title,
			CoverUrl:     v.CoverUrl,
			AuthorId:     v.UserId,
			LikeCount:    int64(v.FavoriteCnt),
			CommentCount: int64(v.CommentCnt),
			PublishTime:  v.CreatedAt.AsTime().Unix(),
		})
	}

	return videoList, nil
}
