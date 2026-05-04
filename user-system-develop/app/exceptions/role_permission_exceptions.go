package exceptions

import "fmt"

type RolePermissionNotFoundError struct {
	RolePermissionID string
}

func (e *RolePermissionNotFoundError) Error() string {
	if e.RolePermissionID != "" {
		return fmt.Sprintf("role-permission with ID %s not found", e.RolePermissionID)
	}
	return "role-permission not found"
}

type PermissionAlreadyAssignedError struct {
	RoleName       string
	PermissionName string
}

func (e *PermissionAlreadyAssignedError) Error() string {
	return fmt.Sprintf("permission %s is already assigned to role %s", e.PermissionName, e.RoleName)
}

type PermissionNotAssignedError struct {
	RoleName       string
	PermissionName string
}

func (e *PermissionNotAssignedError) Error() string {
	return fmt.Sprintf("permission %s is not assigned to role %s", e.PermissionName, e.RoleName)
}
