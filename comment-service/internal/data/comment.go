package data

import (
	pbUser "comment-service/api/user/v1"
	pbVideo "comment-service/api/video/v1"
	"comment-service/internal/biz"
	"comment-service/internal/biz/param"
	"comment-service/internal/data/model"
	"comment-service/internal/data/query"
	middleware "comment-service/internal/pkg/middle"
	"comment-service/internal/pkg/tracing"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type commentRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewCommentRepo(data *Data, logger log.Logger) biz.CommentRepo {
	return &commentRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// ParseToken 解析token
func (c *commentRepo) oldParseToken(ctx context.Context, token, refreshToken string) (int64, error) {

	ctx, span := tracing.StartSpan(ctx, "commentRepo.ParseToken",
		attribute.String("token.length", fmt.Sprintf("%d", len(token))),
	)
	defer span.End()

	resp, err := c.data.UserClient.ParseToken(ctx, &pbUser.ParseTokenRequest{
		Token:        token,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return 0, err
	}

	return resp.UserId, nil
}

var (
	// 全局限流器，10 QPS，burst 20
	commentRateLimiter = middleware.RateLimitMiddleware(10, 20)

	// comment -> user 服务调用熔断器
	userParseTokenCB = middleware.NewCircuitBreaker("user-parse-token")
)

func (c *commentRepo) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {

	exec := func(ctx context.Context) (interface{}, error) {
		ctx, span := tracing.StartSpan(ctx, "commentRepo.ParseToken",
			attribute.String("token.length", fmt.Sprintf("%d", len(token))),
		)
		defer span.End()

		return userParseTokenCB.Execute(func() (interface{}, error) {
			return c.data.UserClient.ParseToken(ctx, &pbUser.ParseTokenRequest{
				Token:        token,
				RefreshToken: refreshToken,
			})
		})
	}

	result, err := commentRateLimiter(func(ctx context.Context, req interface{}) (interface{}, error) {
		return exec(ctx)
	})(ctx, nil)

	if err != nil {
		return 0, err
	}

	resp, ok := result.(*pbUser.ParseTokenReply)
	if !ok || resp == nil {
		return 0, errors.New("failed to parse token")
	}

	return resp.UserId, nil
}

// CreateComment 创建评论
func (c *commentRepo) CreateComment(ctx context.Context, req *param.CreateCommentRequest) (*param.CreateCommentResponse, error) {
	ctx, span := tracing.StartSpan(ctx, "commentRepo.CreateComment",
		attribute.Int64("video_id", req.VideoId),
		attribute.Int64("user_id", req.UserID),
	)
	defer span.End()

	// 1. 生成id
	cid := c.data.idg.Generate()
	err := c.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txQuery := query.Use(tx)

		err := txQuery.Comment.
			WithContext(ctx).
			Create(&model.Comment{
				ID:        cid,
				UserID:    req.UserID,
				VideoID:   req.VideoId,
				ParentID:  req.ParentID,
				Content:   req.Content,
				IsDeleted: false,
			})

		if err != nil {
			return err
		}

		if _, err = txQuery.Video.
			WithContext(ctx).
			Where(txQuery.Video.ID.Eq(req.VideoId)).
			UpdateSimple(txQuery.Video.CommentCnt.Add(1)); err != nil {
			return err
		}
		return nil

	})

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	keyVideoComment := fmt.Sprintf("video:comment:%d", req.VideoId)

	ctxCache, spanCache := tracing.StartSpan(ctx, "Redis.CheckCache")
	if err = c.checkVideoCommentInCache(ctxCache, keyVideoComment, req.VideoId); err != nil {
		spanCache.RecordError(err)
		spanCache.End()
		return nil, err
	}
	spanCache.End()

	// 更新到redis
	ctxIncr, spanIncr := tracing.StartSpan(ctx, "Redis.Incr")
	if err := c.data.rdb.Incr(ctxIncr, keyVideoComment).Err(); err != nil {
		spanIncr.RecordError(err)
		spanIncr.End()
		return nil, err
	}
	spanIncr.End()

	ctxScore, spanScore := tracing.StartSpan(ctx, "UpdateVideoScoreAfterLike")
	err = c.UpdateVideoScoreAfterLike(ctxScore, req.VideoId)
	if err != nil {
		spanScore.RecordError(err)
		spanScore.End()
		return nil, err
	}
	spanScore.End()

	return &param.CreateCommentResponse{
		CommentID: cid,
		Message:   "create comment success",
	}, nil
}

// UpdateVideoScoreAfterLike 更新分数
func (r *commentRepo) UpdateVideoScoreAfterLike(ctx context.Context, videoID int64) error {
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

func (c *commentRepo) checkVideoCommentInCache(ctx context.Context, keyVideoComment string, vid int64) error {
	val, err := c.data.rdb.Exists(ctx, keyVideoComment).Result()
	if err != nil {
		return err
	}

	if val == 0 {
		var cnt int64
		err = c.data.query.
			WithContext(ctx).
			Video.
			Where(query.Video.ID.Eq(vid)).Select(query.Video.CommentCnt).
			Scan(&cnt)
		if err != nil {
			return err
		}
		err = c.data.rdb.Set(ctx, keyVideoComment, cnt, 24*time.Hour).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteComment 删除评论
func (c *commentRepo) DeleteComment(ctx context.Context, commentID, vid int64) error {
	exist, err := c.CheckCommentExist(ctx, commentID)
	if err != nil {
		return err
	}
	if !exist {
		return errors.New("COMMENT_NOT_FOUND ,comment not found or already deleted")
	}
	err = c.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txQuery := query.Use(tx)
		_, err = txQuery.Comment.
			WithContext(ctx).
			Where(c.data.query.Comment.ID.Eq(commentID)).
			Update(c.data.query.Comment.IsDeleted, true)
		if err != nil {
			return err
		}
		if _, err := txQuery.Video.
			WithContext(ctx).
			Where(txQuery.Video.ID.Eq(vid)).
			UpdateSimple(txQuery.Video.CommentCnt.Sub(1)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	keyVideoComment := fmt.Sprintf("video:comment:%d", vid)
	if err = c.checkVideoCommentInCache(ctx, keyVideoComment, vid); err != nil {
		return err
	}

	// 更新到redis
	if err := c.data.rdb.Decr(ctx, keyVideoComment).Err(); err != nil {
		return err
	}

	// 可选: 防止负数出现
	val, err := c.data.rdb.Get(ctx, keyVideoComment).Int64()
	if err != nil {
		return err
	}
	if val < 0 {
		if err := c.data.rdb.Set(ctx, keyVideoComment, 0, 0).Err(); err != nil {
			return err
		}
	}

	err = c.UpdateVideoScoreAfterLike(ctx, vid)
	if err != nil {
		return err
	}

	return nil
}

func (c *commentRepo) CheckVideoExist(ctx context.Context, videoId int64) (bool, error) {
	resp, err := c.data.VideoClient.CheckVideoExists(ctx, &pbVideo.CheckVideoExistsRequest{
		VideoId: videoId,
	})
	if err != nil {
		return false, err
	}
	return resp.Exist, nil
}

func (c *commentRepo) CheckCommentExist(ctx context.Context, commentID int64) (bool, error) {
	_, err := c.data.query.Comment.
		WithContext(ctx).
		Where(c.data.query.Comment.ID.Eq(commentID), c.data.query.Comment.IsDeleted.Is(false)).
		First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *commentRepo) GetCommentList(ctx context.Context, videoId, page, pageSize int64) ([]*model.Comment, error) {
	offset := (page - 1) * pageSize

	comments, err := c.data.query.Comment.
		WithContext(ctx).
		Where(
			c.data.query.Comment.VideoID.Eq(videoId),
			c.data.query.Comment.IsDeleted.Is(false),
		).
		Order(c.data.query.Comment.CreatedAt.Desc()).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find()
	if err != nil {
		return nil, err
	}
	return comments, nil
}
