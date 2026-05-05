package repository

import (
	"morent-backend/internal/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ListAll() ([]models.User, error) {
	var users []models.User
	if findUsersErr := r.db.Find(&users).Error; findUsersErr != nil {
		return nil, findUsersErr
	}
	return users, nil
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	getByEmailErr := r.db.Where("email = ?", email).First(&user).Error
	if getByEmailErr != nil {
		if getByEmailErr == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, getByEmailErr
	}
	return &user, nil
}

func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	getByIDErr := r.db.First(&user, id).Error
	if getByIDErr != nil {
		if getByIDErr == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, getByIDErr
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	result := r.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) Delete(id uint64) error {
	return r.db.Delete(&models.User{}, id).Error
}