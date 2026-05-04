package repositories

import (
	"user-system/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

func (r *UserRoleRepository) GetRolesByUser(userID uuid.UUID) ([]models.Role, error) {
	var userRoles []models.UserRole
	err := r.db.Preload("Role.Permissions").Where("user_id = ?", userID).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}

	roles := make([]models.Role, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = ur.Role
	}
	return roles, nil
}

func (r *UserRoleRepository) Assign(userRole *models.UserRole) error {
	return r.db.Create(userRole).Error
}

func (r *UserRoleRepository) Remove(userID, roleID uuid.UUID) error {
	return r.db.Delete(&models.UserRole{}, "user_id = ? AND role_id = ?", userID, roleID).Error
}
