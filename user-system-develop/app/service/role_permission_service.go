package service

import (
	"user-system/app/dto"

	"github.com/google/uuid"

	"user-system/app/exceptions"
	"user-system/app/models"
	"user-system/app/repositories"
)

type RolePermissionService interface {
	AssignPermissionToRole(roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(roleID, permissionID uuid.UUID) error
	GetRolePermission(id uuid.UUID) (*models.RolePermission, error)
	GetRolePermissions(roleID uuid.UUID) ([]models.RolePermission, error)
	GetPermissionRoles(permissionID uuid.UUID) ([]models.RolePermission, error)
	CheckPermissionAssignment(roleID, permissionID uuid.UUID) (bool, error)
	GetRolesWithPermission(permissionID uuid.UUID) ([]models.Role, error)
	GetPermissionsForRole(roleID uuid.UUID) ([]models.Permission, error)
	BulkAssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	BulkRemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	CheckUserAccess(req dto.CheckUserAccessRequest) (*dto.CheckUserAccessResponse, error)
}

type rolePermissionService struct {
	rolePermissionRepo repositories.RolePermissionRepository
	roleRepo           repositories.RoleRepository
	permissionRepo     repositories.PermissionRepository
	userRepo           repositories.UserRepository
}

func NewRolePermissionService(
	rolePermissionRepo repositories.RolePermissionRepository,
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
) RolePermissionService {
	return &rolePermissionService{
		rolePermissionRepo: rolePermissionRepo,
		roleRepo:           roleRepo,
		permissionRepo:     permissionRepo,
	}
}

func (s *rolePermissionService) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return &exceptions.SystemRoleModificationError{RoleName: role.Name}
	}

	_, err = s.permissionRepo.GetByID(permissionID)
	if err != nil {
		return &exceptions.PermissionNotFoundError{PermissionID: permissionID.String()}
	}

	exists, err := s.rolePermissionRepo.Exists(roleID, permissionID)
	if err != nil {
		return err
	}
	if exists {
		return &exceptions.PermissionAlreadyAssignedError{
			RoleName:       role.Name,
			PermissionName: permissionID.String(),
		}
	}

	rolePermission := &models.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	return s.rolePermissionRepo.Create(rolePermission)
}

func (s *rolePermissionService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return &exceptions.SystemRoleModificationError{RoleName: role.Name}
	}

	exists, err := s.rolePermissionRepo.Exists(roleID, permissionID)
	if err != nil {
		return err
	}
	if !exists {
		return &exceptions.PermissionNotAssignedError{
			RoleName:       role.Name,
			PermissionName: permissionID.String(),
		}
	}

	return s.rolePermissionRepo.DeleteByRoleAndPermission(roleID, permissionID)
}

func (s *rolePermissionService) GetRolePermission(id uuid.UUID) (*models.RolePermission, error) {
	return s.rolePermissionRepo.GetByID(id)
}

func (s *rolePermissionService) GetRolePermissions(roleID uuid.UUID) ([]models.RolePermission, error) {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, err
	}

	return s.rolePermissionRepo.GetByRole(roleID)
}

func (s *rolePermissionService) GetPermissionRoles(permissionID uuid.UUID) ([]models.RolePermission, error) {
	_, err := s.permissionRepo.GetByID(permissionID)
	if err != nil {
		return nil, err
	}

	return s.rolePermissionRepo.GetByPermission(permissionID)
}

func (s *rolePermissionService) CheckPermissionAssignment(roleID, permissionID uuid.UUID) (bool, error) {
	return s.rolePermissionRepo.Exists(roleID, permissionID)
}

func (s *rolePermissionService) GetRolesWithPermission(permissionID uuid.UUID) ([]models.Role, error) {
	_, err := s.permissionRepo.GetByID(permissionID)
	if err != nil {
		return nil, err
	}

	return s.rolePermissionRepo.GetRolesByPermission(permissionID)
}

func (s *rolePermissionService) GetPermissionsForRole(roleID uuid.UUID) ([]models.Permission, error) {
	_, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, err
	}

	return s.rolePermissionRepo.GetPermissionsByRole(roleID)
}

func (s *rolePermissionService) BulkAssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return &exceptions.SystemRoleModificationError{RoleName: role.Name}
	}

	for _, permissionID := range permissionIDs {
		_, err := s.permissionRepo.GetByID(permissionID)
		if err != nil {
			return &exceptions.PermissionNotFoundError{PermissionID: permissionID.String()}
		}

		exists, err := s.rolePermissionRepo.Exists(roleID, permissionID)
		if err != nil {
			return err
		}
		if !exists {
			rolePermission := &models.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			}
			if err := s.rolePermissionRepo.Create(rolePermission); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *rolePermissionService) BulkRemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return &exceptions.SystemRoleModificationError{RoleName: role.Name}
	}

	for _, permissionID := range permissionIDs {
		exists, err := s.rolePermissionRepo.Exists(roleID, permissionID)
		if err != nil {
			return err
		}
		if exists {
			if err := s.rolePermissionRepo.DeleteByRoleAndPermission(roleID, permissionID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *rolePermissionService) CheckUserAccess(req dto.CheckUserAccessRequest) (*dto.CheckUserAccessResponse, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, &exceptions.UserNotFoundError{UserID: req.UserID}
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, &exceptions.UserNotFoundError{UserID: req.UserID}
	}

	if !user.IsActive {
		return &dto.CheckUserAccessResponse{
			HasAccess: false,
			UserID:    userID,
			Resource:  req.ResourceType,
			Action:    req.Action,
			Reason:    "user is inactive",
		}, nil
	}

	permissions, err := s.userRepo.GetUserPermissions(userID)
	if err != nil {
		return nil, err
	}

	requiredPermission := req.Action + "_" + req.ResourceType
	for _, perm := range permissions {
		if perm.Name == requiredPermission || perm.Name == "admin_all" {
			return &dto.CheckUserAccessResponse{
				HasAccess: true,
				UserID:    userID,
				Resource:  req.ResourceType,
				Action:    req.Action,
			}, nil
		}
	}

	return &dto.CheckUserAccessResponse{
		HasAccess: false,
		UserID:    userID,
		Resource:  req.ResourceType,
		Action:    req.Action,
		Reason:    "missing required permission: " + requiredPermission,
	}, nil
}
