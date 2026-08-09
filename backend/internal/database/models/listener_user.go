package models

import "gorm.io/gorm"

type ListenerUser struct {
	gorm.Model
	ListenerID uint   `gorm:"not null;index" json:"listener_id"`
	Username   string `gorm:"not null" json:"username"`
	Password   string `gorm:"not null" json:"password"`
}
