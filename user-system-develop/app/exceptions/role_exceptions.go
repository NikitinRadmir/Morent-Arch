package exceptions

import "fmt"

type RoleAlreadyExistsError struct {
	RoleName string
}

func (e *RoleAlreadyExistsError) Error() string {
	return fmt.Sprintf("role with name %s already exists in this company", e.RoleName)
}

type SystemRoleModificationError struct {
	RoleName string
}

func (e *SystemRoleModificationError) Error() string {
	return fmt.Sprintf("system role %s cannot be modified", e.RoleName)
}

type PermissionNotFoundError struct {
	PermissionID string
}

func (e *PermissionNotFoundError) Error() string {
	return fmt.Sprintf("permission with ID %s not found", e.PermissionID)
}

type UserRoleNotFoundError struct {
	UserID string
	RoleID string
}

func (e *UserRoleNotFoundError) Error() string {
	return "user role not found for user " + e.UserID + " and role " + e.RoleID
}
