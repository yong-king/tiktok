package params

type ListUserVideosRequest struct {
	FUserId  int64
	Page     int32
	PageSize int32
	UserId   int64
}

type ListUserVideosReply struct {
	Videos      []*Video
	Total       int32 // 总视频数
	CurrentPage int32
	PageSize    int32
}

type Video struct {
	Id          int64
	UserId      int64
	PlayUrl     string
	CoverUrl    string
	Title       string
	Description string
	Duration    float32
	Tags        string
	FavoriteCnt int32
	CommentCnt  int32
	ShareCnt    int32
	CollectCnt  int32
}
