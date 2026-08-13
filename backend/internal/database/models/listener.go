package models

import "gorm.io/gorm"

type Listener struct {
	gorm.Model
	Name        string `gorm:"not null" json:"name"`
	Protocol    string `gorm:"type:varchar(50);not null;default:'shadowsocks'" json:"protocol"`
	Type        string `gorm:"type:varchar(50)" json:"type"`
	Port        int    `gorm:"not null" json:"port"`
	BindAddress string `gorm:"type:varchar(100);not null;default:'0.0.0.0'" json:"bind_address"`
	Listen      string `gorm:"type:varchar(100)" json:"listen"`
	TLS         bool   `gorm:"not null;default:false" json:"tls,omitempty"`
	UDP         bool   `gorm:"not null;default:false" json:"udp,omitempty"`
	Enabled     bool   `gorm:"not null;default:false" json:"enabled"`
	Proxy       string `gorm:"type:varchar(255)" json:"proxy,omitempty"`
	Rule        string `gorm:"type:text" json:"rule,omitempty"`
	Config      string `gorm:"type:text" json:"config"`
	Status      string `gorm:"type:varchar(50);default:'inactive'" json:"status"`
}
