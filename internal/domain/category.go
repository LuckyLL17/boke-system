package domain

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Slug      string         `gorm:"size:100;uniqueIndex" json:"slug"`
	ParentID  *uint64        `gorm:"index" json:"parent_id"`
	IconURL   string         `gorm:"size:500" json:"icon_url"`
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Children  []Category     `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}
