package param

type CreateCommentRequest struct {
	ActionType int32
	ParentID   int64
	VideoId    int64
	UserID     int64
	CommentID  int64
	Content    string
}

type CreateCommentResponse struct {
	CommentID int64
	Message   string
}
