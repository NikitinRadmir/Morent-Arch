package models

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"not null;uniqueIndex"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
