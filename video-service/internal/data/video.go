package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"io"
	"math"
	"time"
	pbUser "video-service/api/user/v1"
	v1 "video-service/api/video/v1"
	"video-service/internal/biz/params"
	"video-service/internal/data/model"
	"video-service/internal/pkg/consts"

	"video-service/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type videoRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewVideoRepo(data *Data, logger log.Logger) biz.VideoRepo {
	return &videoRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ParseToken 解析token
func (r *videoRepo) ParseToken(ctx context.Context, token string) (int64, error) {
	claims, err := r.data.jwt.ParseToken(ctx, token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// VerifyAndRefreshTokens 验证token及更新token
func (r *videoRepo) VerifyAndRefreshTokens(ctx context.Context, token string, refreshToken string) (string, error) {
	newToken, err := r.data.jwt.VerifyAndRefreshTokens(ctx, token, refreshToken)
	if err != nil {
		return "", err
	}
	return newToken, nil
}

// UploadVideo 上传视频
func (r *videoRepo) UploadVideo(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	url, err := r.data.uploade.Upload(ctx, objectName, reader, size, contentType)
	if err != nil {
		return "", fmt.Errorf("minio upload failed: %w", err)
	}
	return url, nil
}

// CheckVideoExist 检测视频是否存在
func (r *videoRepo) CheckVideoExist(ctx context.Context, playURL string, userID int64) (bool, error) {
	_, err := r.data.query.Video.
		WithContext(ctx).
		Where(r.data.query.Video.UserID.Eq(userID), r.data.query.Video.PlayURL.Eq(playURL)).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		r.log.WithContext(ctx).Errorf("check video exist error: %v", err)
		return false, err
	}
	return true, nil
}

// CreateVideo 创建视频信息
func (r *videoRepo) CreateVideo(ctx context.Context, in *params.CreateVideoReq) (int64, error) {
	video := &model.Video{
		ID:          r.data.idg.Generate(),
		UserID:      in.UserID,
		PlayURL:     in.PlayUrl,
		CoverURL:    in.CoverUrl,
		Title:       in.Title,
		Description: in.Description,
		Duration:    in.Duration,
		Tags:        in.Tags,
		IsPublic:    in.IsPublic,
		AuditStatus: consts.AuditStatusPending,
		IsOriginal:  in.IsOriginal,
	}

	// **关键补充：保证 BizExt 不为空**
	if video.BizExt == "" {
		video.BizExt = "{}"
	}

	// 1. 保存到sql
	err := r.data.query.Video.
		WithContext(ctx).Omit(r.data.query.Video.DeleteAt).Create(video)
	if err != nil {
		r.log.Errorf("Create video err :%v", err)
		return 0, err
	}

	// 2. 保存到redis
	key := fmt.Sprintf("video:%d", video.ID)
	videoJson, err := json.Marshal(video)
	if err != nil {
		r.log.Errorf("Create video json err :%v", err)
		return video.ID, nil
	}

	err = r.data.rdb.Set(ctx, key, videoJson, 24*time.Hour).Err()
	if err != nil {
		r.log.Errorf("Create video err :%v", err)
	}

	// 初始化分数
	err = r.videoScore(ctx, video.ID)
	if err != nil {
		r.log.Errorf("Create video err :%v", err)
		return 0, err
	}

	return video.ID, nil
}

func (r *videoRepo) videoScore(ctx context.Context, videoID int64) error {
	likeCnt, commentCnt := int64(0), int64(0)
	uploadTime := timestamppb.New(time.Now())
	score := r.CalcVideoScore(ctx, likeCnt, commentCnt, uploadTime)
	err := r.data.rdb.ZAdd(ctx, "video:score", redis.Z{Score: float64(score), Member: videoID}).Err()
	if err != nil {
		r.log.Errorf("Score err: %v", err)
		return err
	}
	return nil
}

// ListUserVideos 根据用户id获取视频列表
func (r *videoRepo) ListUserVideos(ctx context.Context, userID int64, page int32, pageSize int32) ([]*params.Video, int32, error) {
	offset := (page - 1) * pageSize

	db := r.data.query.Video.
		WithContext(ctx).
		Where(r.data.query.Video.UserID.Eq(userID)).
		Order(r.data.query.Video.CreatedAt.Desc(), r.data.query.Video.ID.Desc())

	total, err := db.Count()
	if err != nil {
		r.log.WithContext(ctx).Errorf("list video count err: %v", err)
		return nil, 0, err
	}
	videos, err := db.Offset(int(offset)).Limit(int(pageSize)).Find()
	if err != nil {
		r.log.WithContext(ctx).Errorf("list video find err: %v", err)
		return nil, 0, err
	}

	res := make([]*params.Video, 0, len(videos))
	for _, v := range videos {
		res = append(res, &params.Video{
			Id:          v.ID,
			UserId:      v.UserID,
			PlayUrl:     v.PlayURL,
			CoverUrl:    v.CoverURL,
			Title:       v.Title,
			Description: v.Description,
			Duration:    v.Duration,
			Tags:        v.Tags,
			FavoriteCnt: v.FavoriteCnt,
			CommentCnt:  v.CommentCnt,
			ShareCnt:    v.ShareCnt,
			CollectCnt:  v.CollectCnt,
		})
	}

	return res, int32(total), nil
}

func (r *videoRepo) CheckUserExistByUserID(ctx context.Context, user_id int64) (*pbUser.CheckUserExistByUserIDReply, error) {
	r.log.Infof("CheckUserExistByUserID: user_id: %d", user_id)
	return r.data.UserClient.CheckUserExistByUserID(ctx, &pbUser.CheckUserExistByUserIDRequest{UserId: user_id})
}

func (r *videoRepo) BatchGetVideoInfo(ctx context.Context, ids []int64, page, pageSize int64) ([]*v1.Video, error) {
	//if page <= 0 {
	//	page = 1
	//}
	//if pageSize <= 0 {
	//	pageSize = 10
	//}
	//offset := (page - 1) * pageSize

	videos, err := r.data.query.Video.WithContext(ctx).Where(r.data.query.Video.UserID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	//r.log.Infof("----------BatchGetVideoInfo----------: %v, %d", videos, len(videos))

	if len(videos) == 0 {
		return []*v1.Video{}, nil
	}

	res := make([]*v1.Video, 0, len(videos))
	for _, v := range videos {
		res = append(res, &v1.Video{
			Id:          v.ID,
			UserId:      v.UserID,
			PlayUrl:     v.PlayURL,
			CoverUrl:    v.CoverURL,
			Title:       v.Title,
			FavoriteCnt: v.FavoriteCnt,
			CommentCnt:  v.CommentCnt,
			CreatedAt:   timestamppb.New(v.CreatedAt),
		})
	}
	return res, nil
}

func (r *videoRepo) CheckVideoExistsByID(ctx context.Context, videoID int64) (bool, error) {
	_, err := r.data.query.Video.WithContext(ctx).Where(r.data.query.Video.ID.Eq(videoID)).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		r.log.WithContext(ctx).Errorf("check video exist error: %v", err)
		return false, err
	}
	return true, nil
}

func (r *videoRepo) CalcVideoScore(ctx context.Context, likeCnt int64, commentCnt int64, uploadTime *timestamppb.Timestamp) float64 {
	hours := time.Since(uploadTime.AsTime()).Hours()
	r.log.WithContext(ctx).Infof("CalcVideoScore likeCnt: %d, commentCnt: %d, uploadTime: %s", likeCnt, commentCnt, hours)
	if hours < 0 {
		hours = 0
	}
	timeDecay := 1 / math.Pow(hours+2, 1.2)
	return float64(likeCnt)*1 + float64(commentCnt)*2 + 1000*timeDecay
}

func (r *videoRepo) GetVideoFavoriteAndCommentCount(ctx context.Context, videoID int64) (int64, int64, time.Time, error) {
	r.log.WithContext(ctx).Infof("GetVideoFavoriteAndCommentCount videoID: %d", videoID)
	videoInfo, err := r.data.query.Video.WithContext(ctx).Where(r.data.query.Video.ID.Eq(videoID)).First()
	if err != nil {
		r.log.WithContext(ctx).Errorf("get video err: %v", err)
		return 0, 0, time.Time{}, err
	}

	return int64(videoInfo.FavoriteCnt), int64(videoInfo.CommentCnt), videoInfo.CreatedAt, nil

}

func (r *videoRepo) GetVideoByTitle(ctx context.Context, title string) ([]*v1.Video, error) {
	r.log.WithContext(ctx).Infof("GetVideoByTitle: %v", title)
	r.log.WithContext(ctx).Debugf("esIndex: %s", r.data.esIndex)
	// 1. 在es中模糊查询
	res, err := r.data.es.Search().
		Index(r.data.esIndex).
		Query(
			&types.Query{
				Match: map[string]types.MatchQuery{
					"title": {Query: title},
				},
			},
		).
		Size(20).
		Do(ctx)
	if err != nil {
		r.log.WithContext(ctx).Errorf("get video err: %v", err)
		// 2️⃣ 兜底用 SQL
		return r.getVideoByTitleFromDB(ctx, title)
	}
	if res == nil || res.Hits.Hits == nil || len(res.Hits.Hits) == 0 {
		r.log.WithContext(ctx).Warnf("ES no hits for title: %s, fallback to DB", title)
		return r.getVideoByTitleFromDB(ctx, title)
	}

	// 2. 在sql中模糊查询兜底

	videos := make([]*v1.Video, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		var video v1.Video
		if err := json.Unmarshal(hit.Source_, &video); err != nil {
			r.log.WithContext(ctx).Errorf("unmarshal video err: %v, source: %s", err, string(hit.Source_))
			continue
		}
		videos = append(videos, &video)
	}

	if len(videos) == 0 {
		return r.getVideoByTitleFromDB(ctx, title)
	}

	return videos, nil
}

func (r *videoRepo) getVideoByTitleFromDB(ctx context.Context, title string) ([]*v1.Video, error) {
	r.log.WithContext(ctx).Infof("getVideoByTitleFromDB: %v", title)

	res, err := r.data.query.Video.
		WithContext(ctx).
		Where(r.data.query.Video.Title.Like(fmt.Sprintf("%%%s%%", title))).
		Order(r.data.query.Video.CreatedAt.Desc()).
		Find()
	if err != nil {
		return nil, err
	}

	videos := make([]*v1.Video, 0, len(res))
	for _, v := range res {
		videos = append(videos, &v1.Video{
			Id:          v.ID,
			UserId:      v.UserID,
			PlayUrl:     v.PlayURL,
			CoverUrl:    v.CoverURL,
			Title:       v.Title,
			Description: v.Description,
			Duration:    v.Duration,
			Tags:        v.Tags,
			FavoriteCnt: v.FavoriteCnt,
			CommentCnt:  v.CommentCnt,

			ShareCnt:   v.ShareCnt,
			CollectCnt: v.CollectCnt,
		})
	}
	return videos, nil
}
