package repositories

import (
	"user-system/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetAll() ([]models.Permission, error)
	GetByID(id uuid.UUID) (*models.Permission, error)
	Create(permission *models.Permission) error
	Update(permission *models.Permission) error
	Delete(id uuid.UUID) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetAll() ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.Find(&permissions).Error
	return permissions, err
}

func (r *permissionRepository) GetByID(id uuid.UUID) (*models.Permission, error) {
	var perm models.Permission
	err := r.db.First(&perm, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *permissionRepository) Create(permission *models.Permission) error {
	return r.db.Create(permission).Error
}

func (r *permissionRepository) Update(permission *models.Permission) error {
	return r.db.Save(permission).Error
}

func (r *permissionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Permission{}, "id = ?", id).Error
}
