package models

import "gorm.io/gorm"

type Listener struct {
	gorm.Model
	Name        string `gorm:"not null" json:"name"`
	Protocol    string `gorm:"type:varchar(50);not null;default:'shadowsocks'" json:"protocol"` // shadowsocks, vmess, vless, trojan, hysteria2, tuic
	Type        string `gorm:"type:varchar(50)" json:"type"`             // keep for compat
	Port        int    `gorm:"not null" json:"port"`
	BindAddress string `gorm:"type:varchar(100);not null;default:'0.0.0.0'" json:"bind_address"`
	Listen      string `gorm:"type:varchar(100)" json:"listen"`           // keep for compat
	TLS         bool   `gorm:"not null;default:false" json:"tls"`
	UDP         bool   `gorm:"not null;default:false" json:"udp"`
	Enabled     bool   `gorm:"not null;default:false" json:"enabled"`
	Proxy       string `gorm:"type:varchar(255)" json:"proxy"`
	Rule        string `gorm:"type:text" json:"rule"`
	Config      string `gorm:"type:text" json:"config"` // e.g. password, uuid, flow, cert, etc.
	Status      string `gorm:"type:varchar(50);default:'inactive'" json:"status"`
}
