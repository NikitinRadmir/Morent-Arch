package repositories

import (
	"user-system/app/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll(limit, offset int, companyID uuid.UUID, isActive *bool) ([]models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByEmailAndCompany(email string, companyID uuid.UUID) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	Count(companyID uuid.UUID, isActive *bool) (int64, error)
	CountRecent(days int, companyID uuid.UUID) (int64, error)
	GetUserRoles(userID uuid.UUID) ([]models.Role, error)
	FindByRole(roleID uuid.UUID) ([]models.User, error)
	FindUsersWithRoles(userIDs []uuid.UUID) ([]models.User, error)
	GetUserPermissions(userID uuid.UUID) ([]models.Permission, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(limit, offset int, companyID uuid.UUID, isActive *bool) ([]models.User, error) {
	var users []models.User
	query := r.db.Preload("Company")

	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Company").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Company").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmailAndCompany(email string, companyID uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ? AND company_id = ?", email, companyID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *userRepository) Count(companyID uuid.UUID, isActive *bool) (int64, error) {
	var count int64
	query := r.db.Model(&models.User{})

	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *userRepository) CountRecent(days int, companyID uuid.UUID) (int64, error) {
	var count int64
	since := time.Now().AddDate(0, 0, -days)

	query := r.db.Model(&models.User{}).Where("created_at >= ?", since)

	if companyID != uuid.Nil {
		query = query.Where("company_id = ?", companyID)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *userRepository) GetUserRoles(userID uuid.UUID) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error

	return roles, err
}

func (r *userRepository) FindByRole(roleID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ?", roleID).
		Where("users.deleted_at IS NULL").
		Preload("UserRoles").
		Preload("UserRoles.Role").
		Find(&users).Error
	return users, err
}

func (r *userRepository) FindUsersWithRoles(userIDs []uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Where("id IN ?", userIDs).
		Where("deleted_at IS NULL").
		Preload("UserRoles").
		Preload("UserRoles.Role").
		Find(&users).Error
	return users, err
}

func (r *userRepository) GetUserPermissions(userID uuid.UUID) ([]models.Permission, error) {
	var permissions []models.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct("permissions.*").
		Find(&permissions).Error
	return permissions, err
}
