package domain

import (
	"time"

	"gorm.io/gorm"
)

type Chapter struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	EpisodeID   uint64         `gorm:"index;not null" json:"episode_id"`
	Title       string         `gorm:"size:300;not null" json:"title"`
	StartTime   int            `gorm:"not null" json:"start_time"`
	EndTime     int            `json:"end_time"`
	Description string         `gorm:"type:text" json:"description"`
	ImageURL    string         `gorm:"size:500" json:"image_url"`
	URL         string         `gorm:"size:500" json:"url"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
