package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"not null;uniqueIndex"`
	Description string
	Domain      string `gorm:"uniqueIndex"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	OwnerID     *uuid.UUID
	IsActive    bool `gorm:"default:true"`
}
