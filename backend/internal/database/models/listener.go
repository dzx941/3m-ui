package models

import "gorm.io/gorm"

type Listener struct {
	gorm.Model
	Name    string `gorm:"not null" json:"name"`
	Type    string `gorm:"not null" json:"type"` // mixed, socks, http, redir, tproxy, tunnel, shadowsocks, vmess, tuic
	Listen  string `gorm:"not null" json:"listen"`
	Port    int    `gorm:"not null" json:"port"`
	UDP     bool   `gorm:"not null;default:false" json:"udp"`
	Enabled bool   `gorm:"not null;default:false" json:"enabled"`
	Proxy   string `gorm:"type:varchar(255)" json:"proxy"`
	Rule    string `gorm:"type:text" json:"rule"`
	Config  string `gorm:"type:text" json:"config"` // extra YAML/JSON parameters
}
