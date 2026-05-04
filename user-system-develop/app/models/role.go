package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_role_name_company"`
	Name        string    `gorm:"not null;uniqueIndex:idx_role_name_company"`
	Description string
	IsSystem    bool `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Company Company `gorm:"foreignKey:CompanyID"`

	Permissions []Permission `gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE;"`
}
