package repository

import (
	"time"

	"morent-backend/internal/models"

	"gorm.io/gorm"
)

type RentalRepository struct {
	db *gorm.DB
}

func NewRentalRepository(db *gorm.DB) *RentalRepository {
	return &RentalRepository{db: db}
}

func (r *RentalRepository) ListAll() ([]models.Rental, error) {
	var rentals []models.Rental
	findRentalsErr := r.db.Preload("Car").Order("created_at DESC").Find(&rentals).Error
	return rentals, findRentalsErr
}

func (r *RentalRepository) GetByID(id uint) (*models.Rental, error) {
	var rental models.Rental
	getRentalErr := r.db.Preload("Car").First(&rental, id).Error
	if getRentalErr != nil {
		if getRentalErr == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, getRentalErr
	}
	return &rental, nil
}

func (r *RentalRepository) Create(rental *models.Rental) error {
	return r.db.Create(rental).Error
}

func (r *RentalRepository) ListByUser(userID uint) ([]models.Rental, error) {
	var rentals []models.Rental
	findRentalsErr := r.db.Preload("Car").Where("user_id = ?", userID).Order("created_at DESC").Find(&rentals).Error
	return rentals, findRentalsErr
}

func (r *RentalRepository) ListByCar(carID uint) ([]models.Rental, error) {
	var rentals []models.Rental
	findRentalsErr := r.db.Where("car_id = ?", carID).Order("start_date ASC").Find(&rentals).Error
	return rentals, findRentalsErr
}

func (r *RentalRepository) HasOverlap(carID uint, startDate, endDate time.Time) (bool, error) {
	var count int64
	overlapCountErr := r.db.Model(&models.Rental{}).
		Where("car_id = ?", carID).
		Where("start_date < ? AND end_date > ?", endDate, startDate).
		Count(&count).Error
	return count > 0, overlapCountErr
}

func (r *RentalRepository) HasRental(userID, carID uint) (bool, error) {
	var count int64
	rentalCountErr := r.db.Model(&models.Rental{}).
		Where("user_id = ? AND car_id = ?", userID, carID).
		Count(&count).Error
	return count > 0, rentalCountErr
}

func (r *RentalRepository) Delete(id uint) error {
	return r.db.Delete(&models.Rental{}, id).Error
}
