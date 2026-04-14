package service

import (
	"errors"
	"strconv"

	"morent-backend/internal/models"
)

type CarRepository interface {
	GetAll() ([]models.Car, error)
	GetByID(id int) (*models.Car, error)
	GetFiltered(name, carType string, capacity *int, priceUnder *float64) ([]models.Car, error)
	CreateCar(car *models.Car) error
	UpdateCar(car *models.Car) error
	DeleteCar(id uint) error
}

type CarService struct {
	repo CarRepository
}

func NewCarService(repo CarRepository) *CarService {
	return &CarService{repo: repo}
}

func (s *CarService) GetAll() ([]models.Car, error) {
	return s.repo.GetAll()
}

func (s *CarService) GetByID(id int) (*models.Car, error) {
	return s.repo.GetByID(id)
}

// GetByIDString удобен для GraphQL-слоя, где id приходит строкой.
func (s *CarService) GetByIDString(idStr string) (*models.Car, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}
	return s.GetByID(id)
}

func (s *CarService) GetFiltered(name, carType string, capacity *int, priceUnder *float64) ([]models.Car, error) {
	return s.repo.GetFiltered(name, carType, capacity, priceUnder)
}

func (s *CarService) Create(car *models.Car) error {
	return s.repo.CreateCar(car)
}

func (s *CarService) Update(car *models.Car) error {
	if car.ID == 0 {
		return errors.New("car ID must be provided")
	}
	return s.repo.UpdateCar(car)
}

func (s *CarService) Delete(id uint) error {
	return s.repo.DeleteCar(id)
}

