package models

import (
	"time"

	"gorm.io/gorm"
)

type AccessToken struct {
	gorm.Model
	Name     string     `gorm:"not null" json:"name"`
	Token    string     `gorm:"uniqueIndex;not null" json:"-"`
	Enabled  bool       `gorm:"default:true" json:"enabled"`
	ExpireAt *time.Time `json:"expire_at"`
	Type     string     `gorm:"not null" json:"type"`      // "listener", "user" (legacy), or "proxy" (legacy)
	TargetID uint       `gorm:"not null" json:"target_id"` // Listener ID, ProxyUser ID, or ProxyNode index
}
