package models

import (
	"time"

	"gorm.io/gorm"
)

type AccessToken struct {
	gorm.Model
	Name     string     `gorm:"not null" json:"name"`
	Token    string     `gorm:"uniqueIndex;not null" json:"token"`
	Enabled  bool       `gorm:"default:true" json:"enabled"`
	ExpireAt *time.Time `json:"expire_at"`
	Type     string     `gorm:"not null" json:"type"`      // "user" or "proxy"
	TargetID uint       `gorm:"not null" json:"target_id"` // ProxyUser ID or ProxyNode index
}
