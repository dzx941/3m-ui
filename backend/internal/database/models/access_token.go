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
	Type     string     `gorm:"not null" json:"type"`      // "user" or "proxy"; listener access uses Scope="listener" for compatibility
	Scope    string     `gorm:"type:varchar(32)" json:"scope"` // "listener" for direct Listener client access
	TargetID uint       `gorm:"not null" json:"target_id"` // Listener ID, ProxyUser ID, or ProxyNode index
}
