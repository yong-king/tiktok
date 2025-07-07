package params

type CreateVideoReq struct {
	Title       string
	Description string
	PlayUrl     string
	CoverUrl    string
	Duration    float32
	Tags        string
	IsPublic    bool
	IsOriginal  bool
	SourceUrl   string
	UserID      int64
}

type CreateVideoReply struct {
	VideoId int64
}
