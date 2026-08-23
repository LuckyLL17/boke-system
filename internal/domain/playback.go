package domain

import (
	"time"
)

type Playback struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	EpisodeID    uint64    `gorm:"index;not null" json:"episode_id"`
	ChannelID    uint64    `gorm:"index;not null" json:"channel_id"`
	UserID       uint64    `gorm:"index" json:"user_id"`
	IPAddress    string    `gorm:"size:50;index" json:"ip_address"`
	UserAgent    string    `gorm:"size:1000" json:"user_agent"`
	DeviceType   string    `gorm:"size:50" json:"device_type"`
	OS           string    `gorm:"size:50" json:"os"`
	Browser      string    `gorm:"size:50" json:"browser"`
	Country      string    `gorm:"size:50" json:"country"`
	Region       string    `gorm:"size:50" json:"region"`
	City         string    `gorm:"size:100" json:"city"`
	StartAt      time.Time `gorm:"index" json:"start_at"`
	EndAt        time.Time `json:"end_at"`
	Duration     int       `json:"duration"`
	ListenTime   int       `json:"listen_time"`
	Progress     float64   `gorm:"default:0" json:"progress"`
	Completed    bool      `gorm:"default:false" json:"completed"`
	IsSubscriber bool      `gorm:"default:false" json:"is_subscriber"`
	Source       string    `gorm:"size:50" json:"source"`
}

type StatsCache struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	ChannelID   uint64    `gorm:"index;not null" json:"channel_id"`
	Date        time.Time `gorm:"index" json:"date"`
	PlayCount   int64     `gorm:"default:0" json:"play_count"`
	Listeners   int       `gorm:"default:0" json:"listeners"`
	Subscribers int       `gorm:"default:0" json:"subscribers"`
	Data        string    `gorm:"type:text" json:"data"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
