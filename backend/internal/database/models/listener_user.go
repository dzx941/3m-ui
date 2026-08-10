package models

import "gorm.io/gorm"

// ListenerUser is the explicit many-to-many join between ProxyUser and Listener.
// Username/Password are retained for database compatibility with the Phase 1/2 schema;
// new code uses ProxyUserID and never stores proxy credentials in this join table.
type ListenerUser struct {
	gorm.Model

	ListenerID  uint `gorm:"not null;index;uniqueIndex:idx_listener_proxy_user,priority:1" json:"listener_id"`
	ProxyUserID uint `gorm:"not null;index;uniqueIndex:idx_listener_proxy_user,priority:2" json:"proxy_user_id"`

	// Deprecated legacy fields. They are not used for Phase 6 authentication generation.
	Username string `gorm:"size:100;not null;default:''" json:"-"`
	Password string `gorm:"type:text;not null;default:''" json:"-"`
}
