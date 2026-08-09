package models

import "gorm.io/gorm"

type Config struct {
	gorm.Model
	Name    string `gorm:"uniqueIndex;not null" json:"name"`
	Content string `gorm:"type:text" json:"content"`
}
