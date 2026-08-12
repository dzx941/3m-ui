package models

import (
	"time"

	"gorm.io/gorm"
)

// AccessToken grants a client access to one Listener's generated node
// configuration. Listener is the single source of truth for client exports.
type AccessToken struct {
	gorm.Model
	Name       string     `gorm:"not null" json:"name"`
	Token      string     `gorm:"uniqueIndex;not null" json:"-"`
	Enabled    bool       `gorm:"default:true" json:"enabled"`
	ExpireAt   *time.Time `json:"expire_at"`
	ListenerID uint       `gorm:"not null;index" json:"listener_id"`
}
