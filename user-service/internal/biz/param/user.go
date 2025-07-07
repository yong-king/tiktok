package param

type RegisterParam struct {
	Username string
	Password string
}

type RegisterReplyParam struct {
	Status_code int32
	Status_msg  string
	UserID      int64
	Token       string
}

type LoginParam struct {
	Username string
	Password string
}

type LoginReplyParam struct {
	Status_code  int32
	Status_msg   string
	UserID       int64
	Token        string
	RefreshToken string
}

type UserValidateParam struct {
	Password string
	UserID   int64
}

type UserInfoParam struct {
	ID              int64
	Name            string
	FollowCount     int32
	FollowerCount   int32
	IsFollow        bool
	Avatar          string
	BackgroundImage string
	Signature       string
	TotalFavorited  int32
	WorkCount       int32
	FavoriteCount   int32
}

type UserInfoReplyParam struct {
	Status_code int32
	Status_msg  string
	User        *UserInfoParam
}

type Author struct {
	ID        int64
	Name      string
	AvatarURL string
}
