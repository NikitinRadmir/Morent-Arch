package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateCompanyRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
	Domain      string `json:"domain" binding:"max=100"`
}

type UpdateCompanyRequest struct {
	Name        string `json:"name,omitempty" binding:"max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
	Domain      string `json:"domain,omitempty" binding:"max=100"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

type CompanyResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Domain      string    `json:"domain,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

type CreateCompanyPublicRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
	Domain      string `json:"domain" binding:"max=100"`
}
