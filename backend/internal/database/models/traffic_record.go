package models

import "gorm.io/gorm"

// TrafficRecord stores sampled traffic information for a proxy user.
// It keeps history separate from ProxyUser to avoid growing the user table.
type TrafficRecord struct {
	gorm.Model

	ProxyUserID   uint  `gorm:"index;not null" json:"proxy_user_id"`
	UploadBytes   int64 `gorm:"not null;default:0" json:"upload_bytes"`
	DownloadBytes int64 `gorm:"not null;default:0" json:"download_bytes"`
	Online        bool  `gorm:"default:false" json:"online"`
}
