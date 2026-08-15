package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username           string `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash       string `gorm:"not null" json:"-"`
	Role               string `gorm:"not null" json:"role"` // e.g., "admin", "user"
	MustChangePassword bool   `gorm:"not null;default:false" json:"must_change_password"`
	SessionVersion     uint   `gorm:"not null;default:1" json:"-"`
}
