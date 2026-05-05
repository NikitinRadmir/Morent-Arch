package service

import (
	"errors"
	"time"

	"morent-backend/internal/models"
)

type RentalRepository interface {
	HasOverlap(carID uint, startDate, endDate time.Time) (bool, error)
	Create(rental *models.Rental) error
	ListByUser(userID uint) ([]models.Rental, error)
	ListByCar(carID uint) ([]models.Rental, error)
	GetByID(id uint) (*models.Rental, error)
	Delete(id uint) error
}

type RentalCarRepository interface {
	GetByID(id int) (*models.Car, error)
}

type RentalService struct {
	rentalRepo RentalRepository
	carRepo    RentalCarRepository
}

func NewRentalService(rentalRepo RentalRepository, carRepo RentalCarRepository) *RentalService {
	return &RentalService{
		rentalRepo: rentalRepo,
		carRepo:    carRepo,
	}
}

var ErrInvalidRentalPeriod = errors.New("invalid rental period")
var ErrCarAlreadyBooked = errors.New("car already booked for selected period")
var ErrForbidden = errors.New("forbidden")

func (s *RentalService) CreateRental(userID, carID uint, startDate, endDate time.Time, totalPrice float64) (*models.RentalResponse, error) {
	car, getCarByIDErr := s.carRepo.GetByID(int(carID))
	if getCarByIDErr != nil {
		return nil, getCarByIDErr
	}
	if car == nil {
		return nil, ErrCarNotFound
	}

	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return nil, ErrInvalidRentalPeriod
	}

	hasOverlap, errOverlap := s.rentalRepo.HasOverlap(carID, startDate, endDate)
	if errOverlap != nil {
		return nil, errOverlap
	}
	if hasOverlap {
		return nil, ErrCarAlreadyBooked
	}

	rental := models.Rental{
		UserID:     userID,
		CarID:      carID,
		StartDate:  startDate,
		EndDate:    endDate,
		TotalPrice: totalPrice,
		Car:        *car,
	}

	if errCreate := s.rentalRepo.Create(&rental); errCreate != nil {
		return nil, errCreate
	}

	resp := rental.ToResponse()
	return &resp, nil
}

func (s *RentalService) ListRentals(userID uint) ([]models.RentalResponse, error) {
	rentals, listRentalsErr := s.rentalRepo.ListByUser(userID)
	if listRentalsErr != nil {
		return nil, listRentalsErr
	}

	responses := make([]models.RentalResponse, 0, len(rentals))
	for _, rental := range rentals {
		responses = append(responses, rental.ToResponse())
	}
	return responses, nil
}

func (s *RentalService) ListCarBookings(carID uint) ([]models.Rental, error) {
	return s.rentalRepo.ListByCar(carID)
}

func (s *RentalService) GetRentalByID(id uint) (*models.Rental, error) {
	return s.rentalRepo.GetByID(id)
}

func (s *RentalService) CancelRental(id, userID uint) (bool, error) {
	rental, getRentalErr := s.GetRentalByID(id)
	if getRentalErr != nil {
		return false, getRentalErr
	}
	if rental == nil {
		return false, nil
	}
	if rental.UserID != userID {
		return false, ErrForbidden
	}
	if errDel := s.rentalRepo.Delete(id); errDel != nil {
		return false, errDel
	}
	return true, nil
}
