package models

import "gorm.io/gorm"

type Config struct {
	gorm.Model
	Name    string `gorm:"uniqueIndex;not null" json:"name"`
	Type    string `gorm:"type:varchar(50);not null;default:'custom'" json:"type"` // e.g., "custom", "dns", "general"
	Content string `gorm:"type:text" json:"content"`
	Enabled bool   `gorm:"not null;default:true" json:"enabled"`
}
