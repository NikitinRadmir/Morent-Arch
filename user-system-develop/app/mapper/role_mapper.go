package mapper

import (
	"user-system/app/dto"
	"user-system/app/models"
)

func RoleToDTO(role *models.Role) dto.RoleResponse {
	return dto.RoleResponse{
		ID:          role.ID,
		CompanyID:   role.CompanyID,
		Name:        role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func RolesToDTOs(roles []models.Role) []dto.RoleResponse {
	dtos := make([]dto.RoleResponse, 0, len(roles))
	for _, role := range roles {
		dtos = append(dtos, RoleToDTO(&role))
	}
	return dtos
}

func CreateRequestToModel(req dto.CreateRoleRequest) models.Role {
	return models.Role{
		CompanyID:   req.CompanyID,
		Name:        req.Name,
		Description: req.Description,
	}
}
