package models

import "github.com/google/uuid"

type RolePermission struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoleID       uuid.UUID `gorm:"type:uuid;not null;index"`
	PermissionID uuid.UUID `gorm:"type:uuid;not null;index"`

	Role       Role       `gorm:"foreignKey:RoleID"`
	Permission Permission `gorm:"foreignKey:PermissionID"`
}
