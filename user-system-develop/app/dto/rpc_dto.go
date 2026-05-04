package dto

import (
	"time"

	"github.com/google/uuid"
)

type DepartmentInfo struct {
	Name  string          `json:"name"`
	Users []UserBriefInfo `json:"users"`
	Count int             `json:"count"`
}

type UserBriefInfo struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Position  string    `json:"position"`
	IsActive  bool      `json:"is_active"`
}

type CompanyHierarchyResponse struct {
	CompanyID   uuid.UUID        `json:"company_id"`
	CompanyName string           `json:"company_name"`
	Departments []DepartmentInfo `json:"departments"`
	TotalUsers  int              `json:"total_users"`
}

// User By Role DTOs
type UserByRoleResponse struct {
	RoleID     uuid.UUID         `json:"role_id"`
	RoleName   string            `json:"role_name"`
	Users      []UserWithDetails `json:"users"`
	TotalCount int               `json:"total_count"`
}

type UserWithDetails struct {
	ID         uuid.UUID  `json:"id"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Department string     `json:"department"`
	Position   string     `json:"position"`
	IsActive   bool       `json:"is_active"`
	LastLogin  *time.Time `json:"last_login,omitempty"`
}

// Role Users DTOs
type RoleUsersResponse struct {
	RoleID      uuid.UUID      `json:"role_id"`
	RoleName    string         `json:"role_name"`
	Description string         `json:"description"`
	IsSystem    bool           `json:"is_system"`
	Users       []UserRoleInfo `json:"users"`
	TotalCount  int            `json:"total_count"`
}

type UserRoleInfo struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	AssignedAt time.Time `json:"assigned_at"`
	AssignedBy uuid.UUID `json:"assigned_by"`
}

// Permission Check DTOs
type CheckUserAccessRequest struct {
	UserID       string `json:"user_id"`
	ResourceType string `json:"resource_type"` // user, company, role
	ResourceID   string `json:"resource_id"`
	Action       string `json:"action"` // create, read, update, delete
}

type CheckUserAccessResponse struct {
	HasAccess bool      `json:"has_access"`
	UserID    uuid.UUID `json:"user_id"`
	Resource  string    `json:"resource"`
	Action    string    `json:"action"`
	Reason    string    `json:"reason,omitempty"`
}
