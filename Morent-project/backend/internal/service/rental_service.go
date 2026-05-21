package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
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
	emails     *EmailNotifier
	bank       *BankService
}

func NewRentalService(rentalRepo RentalRepository, carRepo RentalCarRepository, emails *EmailNotifier, bank *BankService) *RentalService {
	return &RentalService{
		rentalRepo: rentalRepo,
		carRepo:    carRepo,
		emails:     emails,
		bank:       bank,
	}
}

var (
	ErrInvalidRentalPeriod   = errors.New("invalid rental period")
	ErrRentalPriceMismatch   = errors.New("rental price does not match server calculation")
	ErrCarAlreadyBooked      = errors.New("car already booked for selected period")
	ErrForbidden             = errors.New("forbidden")
	ErrBankSessionRequired   = errors.New("bank session required for payment")
	ErrInsufficientBankBalance = errors.New("insufficient funds on bank account")
)

func rentalPeriodOverlaps(existing models.Rental, start, end time.Time) bool {
	return start.Before(existing.EndDate) && end.After(existing.StartDate)
}

func (s *RentalService) CreateRental(userID, carID uint, startDate, endDate time.Time, clientTotalPrice float64, bankToken string) (*models.RentalResponse, error) {
	if strings.TrimSpace(bankToken) == "" {
		return nil, ErrBankSessionRequired
	}
	if s.bank == nil || !s.bank.Available() {
		return nil, ErrBankUnavailable
	}

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
	if math.Abs(clientTotalPrice-totalPrice) > 0.01 {
		return nil, ErrRentalPriceMismatch
	}

	existing, err := s.rentalRepo.ListByCar(carID)
	if err != nil {
		return nil, err
	}
	for _, b := range existing {
		if rentalPeriodOverlaps(b, startDate, endDate) {
			return nil, ErrCarAlreadyBooked
		}
	}

	profile, err := s.bank.Profile(bankToken)
	if err != nil {
		return nil, err
	}
	if profile.Balance < totalPrice {
		return nil, ErrInsufficientBankBalance
	}

	idempotencyKey := fmt.Sprintf("rental-%d-%d-%s-%s", userID, carID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if _, err := s.bank.Pay(bankToken, totalPrice, idempotencyKey); err != nil {
		return nil, err
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
		_, _ = s.bank.Deposit(bankToken, totalPrice)
		if errors.Is(errCreate, repository.ErrBookingOverlap) {
			return nil, ErrCarAlreadyBooked
		}
		return nil, errCreate
	}

	resp := rental.ToResponse()
	if s.emails != nil {
		s.emails.NotifyBookingConfirmation(userID, &resp)
	}
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
