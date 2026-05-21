package service

import (
	"errors"
	"math"
	"time"

	"morent-backend/internal/models"
	"morent-backend/internal/repository"
)

type RentalRepository interface {
	CreateIfNoOverlap(rental *models.Rental, startDate, endDate time.Time) error
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

func (s *RentalService) CreateRental(userID, carID uint, startDate, endDate time.Time) (*models.RentalResponse, error) {
	car, err := s.carRepo.GetByID(int(carID))
	if err != nil {
		return nil, err
	}
	if car == nil {
		return nil, ErrCarNotFound
	}

	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return nil, ErrInvalidRentalPeriod
	}

	days := endDate.Sub(startDate).Hours() / 24
	if days < 1 {
		days = 1
	}
	totalPrice := math.Round(car.Price*days*100) / 100
	if totalPrice <= 0 {
		return nil, ErrInvalidRentalPeriod
	}

	rental := models.Rental{
		UserID:     userID,
		CarID:      carID,
		StartDate:  startDate,
		EndDate:    endDate,
		TotalPrice: totalPrice,
		Car:        *car,
	}

	if errCreate := s.rentalRepo.CreateIfNoOverlap(&rental, startDate, endDate); errCreate != nil {
		if errors.Is(errCreate, repository.ErrBookingOverlap) {
			return nil, ErrCarAlreadyBooked
		}
		return nil, errCreate
	}

	resp := rental.ToResponse()
	return &resp, nil
}

func (s *RentalService) ListRentals(userID uint) ([]models.RentalResponse, error) {
	rentals, err := s.rentalRepo.ListByUser(userID)
	if err != nil {
		return nil, err
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
	rental, err := s.GetRentalByID(id)
	if err != nil {
		return false, err
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
