package service

import (
	"user-system/app/dto"
	"strings"

	"github.com/google/uuid"

	"user-system/app/exceptions"
	"user-system/app/models"
	"user-system/app/repositories"
)

type RoleService interface {
	CreateRole(role *models.Role) error
	GetRole(id uuid.UUID) (*models.Role, error)
	GetRoleByNameAndCompany(name string, companyID uuid.UUID) (*models.Role, error)
	GetCompanyRoles(companyID uuid.UUID) ([]models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(id uuid.UUID) error
	AddPermissionToRole(roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(roleID, permissionID uuid.UUID) error
	GetRolePermissions(roleID uuid.UUID) ([]models.Permission, error)
	GetRoleUsers(roleID uuid.UUID) (*dto.RoleUsersResponse, error)
	GetUserRoles(userID uuid.UUID) ([]models.Role, error)
}

type roleService struct {
	roleRepo       repositories.RoleRepository
	companyRepo    repositories.CompanyRepository
	permissionRepo repositories.PermissionRepository
}

func NewRoleService(
	roleRepo repositories.RoleRepository,
	companyRepo repositories.CompanyRepository,
	permissionRepo repositories.PermissionRepository,
) RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		companyRepo:    companyRepo,
		permissionRepo: permissionRepo,
	}
}

func (s *roleService) CreateRole(role *models.Role) error {
	_, err := s.companyRepo.GetByID(role.CompanyID)
	if err != nil {
		return &exceptions.CompanyNotFoundError{CompanyID: role.CompanyID.String()}
	}

	existingRole, err := s.roleRepo.GetByNameAndCompany(role.Name, role.CompanyID)
	if err == nil && existingRole != nil {
		return &exceptions.RoleAlreadyExistsError{RoleName: role.Name}
	}

	if err := s.validateRole(role); err != nil {
		return err
	}

	return s.roleRepo.Create(role)
}

func (s *roleService) GetRole(id uuid.UUID) (*models.Role, error) {
	return s.roleRepo.GetByID(id)
}

func (s *roleService) GetRoleByNameAndCompany(name string, companyID uuid.UUID) (*models.Role, error) {
	return s.roleRepo.GetByNameAndCompany(name, companyID)
}

func (s *roleService) GetCompanyRoles(companyID uuid.UUID) ([]models.Role, error) {
	_, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, &exceptions.CompanyNotFoundError{CompanyID: companyID.String()}
	}

	return s.roleRepo.GetByCompany(companyID)
}

func (s *roleService) UpdateRole(role *models.Role) error {
	existingRole, err := s.roleRepo.GetByID(role.ID)
	if err != nil {
		return err
	}

	if existingRole.IsSystem {
		return &exceptions.SystemRoleModificationError{RoleName: existingRole.Name}
	}

	if existingRole.Name != role.Name {
		roleWithSameName, err := s.roleRepo.GetByNameAndCompany(role.Name, role.CompanyID)
		if err == nil && roleWithSameName != nil && roleWithSameName.ID != role.ID {
			return &exceptions.RoleAlreadyExistsError{RoleName: role.Name}
		}
	}

	if err := s.validateRole(role); err != nil {
		return err
	}

	return s.roleRepo.Update(role)
}

func (s *roleService) DeleteRole(id uuid.UUID) error {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return &exceptions.SystemRoleModificationError{RoleName: role.Name}
	}

	return s.roleRepo.Delete(id)
}

func (s *roleService) AddPermissionToRole(roleID, permissionID uuid.UUID) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}

	_, err = s.permissionRepo.GetByID(permissionID)
	if err != nil {
		return &exceptions.PermissionNotFoundError{PermissionID: permissionID.String()}
	}

	return s.roleRepo.AddPermission(roleID, permissionID)
}

func (s *roleService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}

	return s.roleRepo.RemovePermission(roleID, permissionID)
}

func (s *roleService) GetRolePermissions(roleID uuid.UUID) ([]models.Permission, error) {
	return s.roleRepo.GetPermissions(roleID)
}

func (s *roleService) validateRole(role *models.Role) error {
	if strings.TrimSpace(role.Name) == "" {
		return &exceptions.ValidationError{Field: "name", Reason: "role name cannot be empty"}
	}

	if len(role.Name) > 100 {
		return &exceptions.ValidationError{Field: "name", Reason: "role name cannot exceed 100 characters"}
	}

	if len(role.Description) > 500 {
		return &exceptions.ValidationError{Field: "description", Reason: "role description cannot exceed 500 characters"}
	}

	return nil
}

func (s *roleService) GetRoleUsers(roleID uuid.UUID) (*dto.RoleUsersResponse, error) {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, &exceptions.RoleNotFoundError{RoleName: roleID.String()}
	}

	userRoles, err := s.roleRepo.GetRoleWithUsers(roleID)
	if err != nil {
		return nil, err
	}

	var users []dto.UserRoleInfo
	for _, ur := range userRoles {
		users = append(users, dto.UserRoleInfo{
			ID:         ur.User.ID,
			Email:      ur.User.Email,
			FirstName:  ur.User.FirstName,
			LastName:   ur.User.LastName,
			AssignedAt: ur.AssignedAt,
			AssignedBy: ur.AssignedBy,
		})
	}

	return &dto.RoleUsersResponse{
		RoleID:      role.ID,
		RoleName:    role.Name,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Users:       users,
		TotalCount:  len(users),
	}, nil
}

func (s *roleService) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	// Получаем все связи пользователь-роль
	userRoles, err := s.roleRepo.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	// Извлекаем роли
	var roles []models.Role
	for _, ur := range userRoles {
		role, err := s.roleRepo.GetByID(ur.RoleID)
		if err != nil {
			continue
		}
		roles = append(roles, *role)
	}

	return roles, nil
}
