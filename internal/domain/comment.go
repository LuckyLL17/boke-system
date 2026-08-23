package domain

import (
	"time"

	"gorm.io/gorm"
)

type CommentStatus int

const (
	CommentPending  CommentStatus = 0
	CommentApproved CommentStatus = 1
	CommentRejected CommentStatus = 2
	CommentReported CommentStatus = 3
)

type CommentType string

const (
	CommentTypeEpisode CommentType = "episode"
	CommentTypeChannel CommentType = "channel"
	CommentTypeBoard   CommentType = "board"
)

type Comment struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	EpisodeID  uint64         `gorm:"index" json:"episode_id"`
	ChannelID  uint64         `gorm:"index" json:"channel_id"`
	UserID     uint64         `gorm:"index;not null" json:"user_id"`
	ParentID   uint64         `gorm:"index" json:"parent_id"`
	Type       CommentType    `gorm:"size:20;default:episode" json:"type"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	Rating     int            `gorm:"default:0" json:"rating"`
	Status     CommentStatus  `gorm:"default:0;index" json:"status"`
	IPAddress  string         `gorm:"size:50" json:"ip_address"`
	UserAgent  string         `gorm:"size:500" json:"user_agent"`
	IsDanmaku  bool           `gorm:"default:false" json:"is_danmaku"`
	DanmakuTime int           `json:"danmaku_time"`
	ReportCount int           `gorm:"default:0" json:"report_count"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	User       *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Episode    *Episode       `gorm:"foreignKey:EpisodeID" json:"episode,omitempty"`
}
