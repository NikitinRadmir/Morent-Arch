package repository

import (
	"morent-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FavoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

func (r *FavoriteRepository) ListAll() ([]models.Favorite, error) {
	var favorites []models.Favorite
	findFavoritesErr := r.db.Preload("Car").Find(&favorites).Error
	return favorites, findFavoritesErr
}

func (r *FavoriteRepository) ListByUser(userID uint) ([]models.Favorite, error) {
	var favorites []models.Favorite
	findFavoritesErr := r.db.Preload("Car").Where("user_id = ?", userID).Find(&favorites).Error
	return favorites, findFavoritesErr
}

func (r *FavoriteRepository) Exists(userID, carID uint) (bool, error) {
	var count int64
	existsCountErr := r.db.Model(&models.Favorite{}).Where("user_id = ? AND car_id = ?", userID, carID).Count(&count).Error
	return count > 0, existsCountErr
}

func (r *FavoriteRepository) Create(userID, carID uint) error {
	favorite := models.Favorite{
		UserID: userID,
		CarID:  carID,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "car_id"}},
		DoNothing: true,
	}).Create(&favorite).Error
}

func (r *FavoriteRepository) Delete(userID, carID uint) error {
	// Используем hard delete, чтобы запись реально удалялась из БД.
	return r.db.Unscoped().Where("user_id = ? AND car_id = ?", userID, carID).Delete(&models.Favorite{}).Error
}

// Remove — алиас для Delete, удобен в админке.
func (r *FavoriteRepository) Remove(userID, carID uint) error {
	return r.Delete(userID, carID)
}
