package dto

type StatsQuery struct {
	ChannelID uint64 `uri:"id"`
	Range     string `form:"range"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

type PlaybackStartDTO struct {
	EpisodeID uint64 `json:"episode_id" binding:"required"`
	Source    string `json:"source" binding:"max=50"`
}

type PlaybackEndDTO struct {
	PlaybackID uint64  `json:"playback_id" binding:"required"`
	ListenTime int     `json:"listen_time"`
	Progress   float64 `json:"progress" binding:"min=0,max=1"`
	Completed  bool    `json:"completed"`
}

type PasswordChangeDTO struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ProfileUpdateDTO struct {
	Nickname  string `json:"nickname" binding:"max=50"`
	AvatarURL string `json:"avatar_url" binding:"max=500"`
}

type LoginDTO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterDTO struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"max=50"`
}
