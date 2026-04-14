package models

import "gorm.io/gorm"

type Session struct {
	gorm.Model
	UserID uint   `gorm:"not null;index"`
	Token  string `gorm:"size:255;uniqueIndex;not null"`
	User   User   `gorm:"constraint:OnDelete:CASCADE"`
}
