package repository

import (
	"morent-backend/internal/models"

	"gorm.io/gorm"
    "strings"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) ListAll() ([]models.Favorite, error) {
	var favorites []models.Favorite
	err := r.db.Preload("Car").Find(&favorites).Error
	return favorites, err
}

func (r *FavoriteRepository) ListByUser(userID uint) ([]models.Favorite, error) {
	var favorites []models.Favorite
	err := r.db.Preload("Car").Where("user_id = ?", userID).Find(&favorites).Error
	return favorites, err
}

func (r *FavoriteRepository) Exists(userID, carID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Favorite{}).Where("user_id = ? AND car_id = ?", userID, carID).Count(&count).Error
	return count > 0, err
}

func (r *FavoriteRepository) Create(userID, carID uint) error {
	favorite := models.Favorite{
		UserID: userID,
		CarID:  carID,
	}
	err := r.db.Where("user_id = ? AND car_id = ?", userID, carID).FirstOrCreate(&favorite).Error
	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil // already exists, считаем успех
	}
	return err
}

func (r *FavoriteRepository) Delete(userID, carID uint) error {
	return r.db.Where("user_id = ? AND car_id = ?", userID, carID).Delete(&models.Favorite{}).Error
}

// Remove — алиас для Delete, удобен в админке.
func (r *FavoriteRepository) Remove(userID, carID uint) error {
	return r.Delete(userID, carID)
}
