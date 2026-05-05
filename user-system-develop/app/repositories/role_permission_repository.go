package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"user-system/app/exceptions"
	"user-system/app/models"
)

type RolePermissionRepository interface {
	Create(rolePermission *models.RolePermission) error
	GetByID(id uuid.UUID) (*models.RolePermission, error)
	GetByRoleAndPermission(roleID, permissionID uuid.UUID) (*models.RolePermission, error)
	GetByRole(roleID uuid.UUID) ([]models.RolePermission, error)
	GetByPermission(permissionID uuid.UUID) ([]models.RolePermission, error)
	Delete(id uuid.UUID) error
	DeleteByRoleAndPermission(roleID, permissionID uuid.UUID) error
	Exists(roleID, permissionID uuid.UUID) (bool, error)
	GetRolesByPermission(permissionID uuid.UUID) ([]models.Role, error)
	GetPermissionsByRole(roleID uuid.UUID) ([]models.Permission, error)
}

type rolePermissionRepository struct {
	db *gorm.DB
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

func (r *rolePermissionRepository) Create(rolePermission *models.RolePermission) error {
	return r.db.Create(rolePermission).Error
}

func (r *rolePermissionRepository) GetByID(id uuid.UUID) (*models.RolePermission, error) {
	var rolePermission models.RolePermission
	err := r.db.Preload("Role").Preload("Permission").
		First(&rolePermission, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &exceptions.RolePermissionNotFoundError{RolePermissionID: id.String()}
		}
		return nil, err
	}
	return &rolePermission, nil
}

func (r *rolePermissionRepository) GetByRoleAndPermission(roleID, permissionID uuid.UUID) (*models.RolePermission, error) {
	var rolePermission models.RolePermission
	err := r.db.Preload("Role").Preload("Permission").
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		First(&rolePermission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &exceptions.RolePermissionNotFoundError{}
		}
		return nil, err
	}
	return &rolePermission, nil
}

func (r *rolePermissionRepository) GetByRole(roleID uuid.UUID) ([]models.RolePermission, error) {
	var rolePermissions []models.RolePermission
	err := r.db.Preload("Permission").
		Where("role_id = ?", roleID).
		Find(&rolePermissions).Error
	if err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepository) GetByPermission(permissionID uuid.UUID) ([]models.RolePermission, error) {
	var rolePermissions []models.RolePermission
	err := r.db.Preload("Role").
		Where("permission_id = ?", permissionID).
		Find(&rolePermissions).Error
	if err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *rolePermissionRepository) Delete(id uuid.UUID) error {
	result := r.db.Where("id = ?", id).Delete(&models.RolePermission{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &exceptions.RolePermissionNotFoundError{RolePermissionID: id.String()}
	}
	return nil
}

func (r *rolePermissionRepository) DeleteByRoleAndPermission(roleID, permissionID uuid.UUID) error {
	result := r.db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&models.RolePermission{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return &exceptions.RolePermissionNotFoundError{}
	}
	return nil
}

func (r *rolePermissionRepository) Exists(roleID, permissionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.RolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *rolePermissionRepository) GetRolesByPermission(permissionID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("role_permissions.permission_id = ?", permissionID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *rolePermissionRepository) GetPermissionsByRole(roleID uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
