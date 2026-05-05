package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
}

type UpdatePermissionRequest struct {
	Name        string `json:"name,omitempty" binding:"max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"company_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RolePermissionResponse struct {
	ID         uuid.UUID          `json:"id"`
	Role       RoleResponse       `json:"role"`
	Permission PermissionResponse `json:"permission"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type CreateRoleRequest struct {
	CompanyID   uuid.UUID `json:"company_id" binding:"required"`
	Name        string    `json:"name" binding:"required,min=1,max=100"`
	Description string    `json:"description,omitempty" binding:"max=500"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name,omitempty" binding:"max=100"`
	Description string `json:"description,omitempty" binding:"max=500"`
}
