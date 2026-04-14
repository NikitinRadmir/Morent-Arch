package models

import "gorm.io/gorm"

type Favorite struct {
	gorm.Model
	UserID uint `gorm:"not null;uniqueIndex:idx_user_car"`
	CarID  uint `gorm:"not null;uniqueIndex:idx_user_car"`
	Car    Car  `gorm:"constraint:OnDelete:CASCADE"`
}
