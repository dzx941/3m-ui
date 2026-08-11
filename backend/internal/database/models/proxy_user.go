package models

import (
	"time"

	"gorm.io/gorm"
)

// ProxyUser is a user account used to authenticate against Mihomo server nodes.
// It is deliberately separate from User, which represents a 3m-ui administrator.
type ProxyUser struct {
	gorm.Model

	Username string `gorm:"size:100;not null;uniqueIndex" json:"username"`
	// PasswordEncrypted contains an AES-GCM encrypted password. It is never serialized to API responses.
	PasswordEncrypted string `gorm:"type:text" json:"-"`
	UUID              string `gorm:"size:64;not null;uniqueIndex" json:"-"`
	TrafficLimit      int64  `gorm:"not null;default:0" json:"traffic_limit"`
	TrafficUsed       int64  `gorm:"not null;default:0" json:"traffic_used"`
	// UploadBytes/DownloadBytes are cumulative counters split by direction.
	// TrafficUsed (above) remains the single source of truth for quota
	// enforcement and equals UploadBytes+DownloadBytes over time.
	UploadBytes   int64 `gorm:"not null;default:0" json:"upload_bytes"`
	DownloadBytes int64 `gorm:"not null;default:0" json:"download_bytes"`
	// LastSeen is updated whenever the traffic collector observes an active
	// Mihomo connection attributable to this user. Nil means never seen.
	LastSeen *time.Time `json:"last_seen"`
	// Online reflects whether the user had at least one active connection
	// during the most recent traffic collection tick.
	Online     bool      `gorm:"not null;default:false" json:"online"`
	ExpireTime time.Time `json:"expire_time"`
	Enabled    bool      `gorm:"not null;default:true" json:"enabled"`
}
