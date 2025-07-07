package data

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
	"io"
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

	err := r.data.query.Video.
		WithContext(ctx).Omit(r.data.query.Video.DeleteAt).Create(video)
	if err != nil {
		r.log.Errorf("Create video err :%v", err)
		return 0, err
	}
	return video.ID, nil
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
