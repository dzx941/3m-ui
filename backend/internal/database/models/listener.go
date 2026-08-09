package models

import "gorm.io/gorm"

type Listener struct {
	gorm.Model
	Name    string `gorm:"not null" json:"name"`
	Type    string `gorm:"not null" json:"type"` // e.g., "shadowsocks", "vmess", "trojan"
	Listen  string `gorm:"not null" json:"listen"`
	Port    int    `gorm:"not null" json:"port"`
	Enabled bool   `gorm:"not null;default:false" json:"enabled"`
	Config  string `gorm:"type:text" json:"config"` // YAML or JSON configuration detail
}
