package repository

import (
	"morent-backend/internal/models"

	"gorm.io/gorm"
)

type CarRepository struct {
	db *gorm.DB
}

func NewCarRepository(db *gorm.DB) *CarRepository {
	return &CarRepository{db: db}
}

func (r *CarRepository) GetAll() ([]models.Car, error) {
	var cars []models.Car
	findCarsErr := r.db.Find(&cars).Error
	return cars, findCarsErr
}

func (r *CarRepository) GetByID(id int) (*models.Car, error) {
	var car models.Car
	getCarErr := r.db.First(&car, id).Error
	if getCarErr == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if getCarErr != nil {
		return nil, getCarErr
	}
	return &car, nil
}

func (r *CarRepository) GetFiltered(name, carType string, capacity *int, priceUnder *float64) ([]models.Car, error) {
	var cars []models.Car
	query := r.db.Model(&models.Car{})

	if name != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%")
	}
	if carType != "" {
		query = query.Where("type = ?", carType)
	}
	if capacity != nil {
		query = query.Where("capacity = ?", *capacity)
	}
	if priceUnder != nil {
		query = query.Where("price <= ?", *priceUnder)
	}

	findFilteredCarsErr := query.Find(&cars).Error
	return cars, findFilteredCarsErr
}

func (r *CarRepository) CreateCar(car *models.Car) error {
	return r.db.Create(car).Error
}

func (r *CarRepository) UpdateCar(car *models.Car) error {
	result := r.db.Model(&models.Car{}).Where("id = ?", car.ID).Updates(car)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CarRepository) DeleteCar(id uint) error {
	result := r.db.Delete(&models.Car{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
