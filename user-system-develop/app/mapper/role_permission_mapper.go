package mapper

import (
	"user-system/app/dto"
	"user-system/app/models"

	"github.com/google/uuid"
)

func PermissionToDTO(permission *models.Permission) dto.PermissionResponse {
	return dto.PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}

func PermissionsToDTOs(permissions []models.Permission) []dto.PermissionResponse {
	dtos := make([]dto.PermissionResponse, 0, len(permissions))
	for _, perm := range permissions {
		dtos = append(dtos, PermissionToDTO(&perm))
	}
	return dtos
}

func RolePermissionToAssignmentDTO(roleID, permissionID uuid.UUID) dto.RolePermissionAssignmentResponse {
	return dto.RolePermissionAssignmentResponse{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
}

func RolePermissionToDetailDTO(rp *models.RolePermission) dto.RolePermissionResponse {
	return dto.RolePermissionResponse{
		ID:         rp.ID,
		Role:       RoleToDTO(&rp.Role),
		Permission: PermissionToDTO(&rp.Permission),
	}
}

func RoleWithPermissionsToDTO(role *models.Role, permissions []models.Permission) dto.RoleWithPermissionsResponse {
	return dto.RoleWithPermissionsResponse{
		Role:        RoleToDTO(role),
		Permissions: PermissionsToDTOs(permissions),
	}
}

func PermissionWithRolesToDTO(permission *models.Permission, roles []models.Role) dto.PermissionWithRolesResponse {
	return dto.PermissionWithRolesResponse{
		Permission: PermissionToDTO(permission),
		Roles:      RolesToDTOs(roles),
	}
}
