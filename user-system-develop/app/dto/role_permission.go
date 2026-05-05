package dto

import (
	"time"

	"github.com/google/uuid"
)

type AssignPermissionRequest struct {
	RoleID       string `json:"role_id" binding:"required,uuid"`
	PermissionID string `json:"permission_id" binding:"required,uuid"`
}

type BulkPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids" binding:"required,min=1,dive,uuid"`
}

type CheckPermissionResponse struct {
	Assigned bool `json:"assigned"`
}

type RolePermissionAssignmentResponse struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
}

type RolePermissionDetailResponse struct {
	ID           uuid.UUID `json:"id"`
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RoleWithPermissionsResponse struct {
	Role        RoleResponse         `json:"role"`
	Permissions []PermissionResponse `json:"permissions"`
}

type PermissionWithRolesResponse struct {
	Permission PermissionResponse `json:"permission"`
	Roles      []RoleResponse     `json:"roles"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
