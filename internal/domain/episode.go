package domain

import (
	"time"

	"gorm.io/gorm"
)

type EpisodeStatus string

const (
	EpisodeDraft     EpisodeStatus = "draft"
	EpisodeScheduled EpisodeStatus = "scheduled"
	EpisodePublished EpisodeStatus = "published"
	EpisodeArchived  EpisodeStatus = "archived"
)

type Episode struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	ChannelID     uint64         `gorm:"index;not null" json:"channel_id"`
	Title         string         `gorm:"size:300;not null" json:"title"`
	Slug          string         `gorm:"size:150;index" json:"slug"`
	EpisodeNumber int            `gorm:"index" json:"episode_number"`
	SeasonNumber  int            `json:"season_number"`
	Description   string         `gorm:"type:text" json:"description"`
	ShowNotes     string         `gorm:"type:text" json:"show_notes"`
	CoverImageURL string         `gorm:"size:500" json:"cover_image_url"`
	AudioFileURL  string         `gorm:"size:500;not null" json:"audio_file_url"`
	AudioFileSize int64          `json:"audio_file_size"`
	Duration      int            `json:"duration"`
	SampleRate    int            `json:"sample_rate"`
	BitRate       int            `json:"bit_rate"`
	MimeType      string         `gorm:"size:50" json:"mime_type"`
	Status        EpisodeStatus  `gorm:"size:20;default:draft;index" json:"status"`
	ScheduledAt   *time.Time     `json:"scheduled_at"`
	PublishedAt   *time.Time     `json:"published_at"`
	Explicit      bool           `gorm:"default:false" json:"explicit"`
	Author        string         `gorm:"size:100" json:"author"`
	CategoryID    *uint64        `json:"category_id"`
	PlayCount     int64          `gorm:"default:0;index" json:"play_count"`
	LikeCount     int            `gorm:"default:0" json:"like_count"`
	CommentCount  int            `gorm:"default:0" json:"comment_count"`
	Tags          string         `gorm:"size:500" json:"tags"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Channel       *Channel       `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Chapters      []Chapter      `gorm:"foreignKey:EpisodeID" json:"chapters,omitempty"`
	Comments      []Comment      `gorm:"foreignKey:EpisodeID" json:"comments,omitempty"`
}
