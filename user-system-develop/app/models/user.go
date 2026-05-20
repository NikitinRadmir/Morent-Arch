package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	MorentUserID *uint     `gorm:"uniqueIndex"`
	CompanyID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Email        string    `gorm:"not null;index"`
	Password   string    `gorm:"not null"`
	FirstName  string
	LastName   string
	Phone      string
	Department string
	Position   string
	IsActive   bool `gorm:"default:true"`
	LastLogin  *time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Company   Company `gorm:"foreignKey:CompanyID"`
	UserRoles []UserRole
}
