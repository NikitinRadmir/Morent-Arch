package repository

import (
	"errors"
	"time"

	"morent-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrBookingOverlap = errors.New("booking overlap")

type RentalRepository struct {
	db *gorm.DB
}

func NewRentalRepository(db *gorm.DB) *RentalRepository {
	return &RentalRepository{db: db}
}

func (r *RentalRepository) ListAll() ([]models.Rental, error) {
	var rentals []models.Rental
	err := r.db.Preload("Car").Order("created_at DESC").Find(&rentals).Error
	return rentals, err
}

func (r *RentalRepository) GetByID(id uint) (*models.Rental, error) {
	var rental models.Rental
	err := r.db.Preload("Car").First(&rental, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rental, nil
}

// CreateIfNoOverlap атомарно проверяет пересечение дат и создаёт бронь.
func (r *RentalRepository) CreateIfNoOverlap(rental *models.Rental, startDate, endDate time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var overlap models.Rental
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("car_id = ?", rental.CarID).
			Where("start_date < ? AND end_date > ?", endDate, startDate).
			Limit(1).
			Take(&overlap).Error
		if err == nil {
			return ErrBookingOverlap
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return tx.Create(rental).Error
	})
}

func (r *RentalRepository) Create(rental *models.Rental) error {
	return r.db.Create(rental).Error
}

func (r *RentalRepository) ListByUser(userID uint) ([]models.Rental, error) {
	var rentals []models.Rental
	err := r.db.Preload("Car").Where("user_id = ?", userID).Order("created_at DESC").Find(&rentals).Error
	return rentals, err
}

func (r *RentalRepository) ListByCar(carID uint) ([]models.Rental, error) {
	var rentals []models.Rental
	err := r.db.Where("car_id = ?", carID).Order("start_date ASC").Find(&rentals).Error
	return rentals, err
}

func (r *RentalRepository) HasOverlap(carID uint, startDate, endDate time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&models.Rental{}).
		Where("car_id = ?", carID).
		Where("start_date < ? AND end_date > ?", endDate, startDate).
		Count(&count).Error
	return count > 0, err
}

func (r *RentalRepository) HasRental(userID, carID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Rental{}).
		Where("user_id = ? AND car_id = ?", userID, carID).
		Count(&count).Error
	return count > 0, err
}

func (r *RentalRepository) Delete(id uint) error {
	return r.db.Delete(&models.Rental{}, id).Error
}
