package domain

import (
	"time"
)

type Subscriber struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	ChannelID  uint64    `gorm:"index;not null" json:"channel_id"`
	Email      string    `gorm:"size:200;index" json:"email"`
	Source     string    `gorm:"size:50" json:"source"`
	IPAddress  string    `gorm:"size:50" json:"ip_address"`
	UserAgent  string    `gorm:"size:1000" json:"user_agent"`
	Country    string    `gorm:"size:50" json:"country"`
	Status     int       `gorm:"default:1;index" json:"status"`
	UnsubToken string    `gorm:"size:100" json:"unsub_token"`
	SubscribedAt time.Time `json:"subscribed_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
