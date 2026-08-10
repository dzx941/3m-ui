package models

import (
	"time"

	"gorm.io/gorm"
)

type Subscription struct {
	gorm.Model
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	Token      string     `gorm:"uniqueIndex;not null" json:"token"`
	Format     string     `gorm:"not null" json:"format"` // e.g., "clash", "sing-box"
	ExpireTime *time.Time `json:"expire_time"`
}
