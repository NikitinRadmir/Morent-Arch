package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	RoleID     uuid.UUID `gorm:"type:uuid;not null;index"`
	AssignedBy uuid.UUID `gorm:"type:uuid"`
	AssignedAt time.Time

	User User `gorm:"foreignKey:UserID"`
	Role Role `gorm:"foreignKey:RoleID"`
}
