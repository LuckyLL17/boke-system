package domain

import (
	"time"

	"gorm.io/gorm"
)

type ChannelStatus int

const (
	ChannelPending  ChannelStatus = 0
	ChannelApproved ChannelStatus = 1
	ChannelRejected ChannelStatus = 2
	ChannelDisabled ChannelStatus = 3
)

type Channel struct {
	ID             uint64         `gorm:"primaryKey" json:"id"`
	OwnerID        uint64         `gorm:"index;not null" json:"owner_id"`
	Title          string         `gorm:"size:200;not null" json:"title"`
	Slug           string         `gorm:"size:100;uniqueIndex" json:"slug"`
	Description    string         `gorm:"type:text" json:"description"`
	CoverImageURL  string         `gorm:"size:500" json:"cover_image_url"`
	Language       string         `gorm:"size:20;default:zh-CN" json:"language"`
	Copyright      string         `gorm:"size:500" json:"copyright"`
	Author         string         `gorm:"size:100" json:"author"`
	Email          string         `gorm:"size:100" json:"email"`
	CategoryID     *uint64        `gorm:"index" json:"category_id"`
	Explicit       bool           `gorm:"default:false" json:"explicit"`
	CustomDomain   string         `gorm:"size:200" json:"custom_domain"`
	ITunesCategory string         `gorm:"size:100" json:"itunes_category"`
	Status         ChannelStatus  `gorm:"default:0;index" json:"status"`
	Subscribers    int            `gorm:"default:0" json:"subscribers"`
	TotalPlays     int64          `gorm:"default:0" json:"total_plays"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Owner          *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Category       *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Episodes       []Episode      `gorm:"foreignKey:ChannelID" json:"episodes,omitempty"`
}
