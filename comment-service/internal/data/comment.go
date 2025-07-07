package data

import (
	pbUser "comment-service/api/user/v1"
	pbVideo "comment-service/api/video/v1"
	"comment-service/internal/biz"
	"comment-service/internal/biz/param"
	"comment-service/internal/data/model"
	"comment-service/internal/data/query"
	"context"
	"errors"
	"gorm.io/gorm"

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
func (c *commentRepo) ParseToken(ctx context.Context, token, refreshToken string) (int64, error) {
	resp, err := c.data.UserClient.ParseToken(ctx, &pbUser.ParseTokenRequest{
		Token:        token,
		RefreshToken: refreshToken,
	})
	if err != nil {
		return 0, err
	}

	return resp.UserId, nil
}

// CreateComment 创建评论
func (c *commentRepo) CreateComment(ctx context.Context, req *param.CreateCommentRequest) (*param.CreateCommentResponse, error) {
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
		return nil, err
	}

	return &param.CreateCommentResponse{
		CommentID: cid,
		Message:   "create comment success",
	}, nil
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
	return c.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

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
