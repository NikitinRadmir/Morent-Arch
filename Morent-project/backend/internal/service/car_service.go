package service

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"morent-backend/internal/cache"
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
	repo  CarRepository
	cache *cache.CarCache
	log   *slog.Logger
}

func NewCarService(repo CarRepository, carCache *cache.CarCache, log *slog.Logger) *CarService {
	if log == nil {
		log = slog.Default()
	}
	return &CarService{repo: repo, cache: carCache, log: log}
}

func (s *CarService) GetAll() ([]models.Car, error) {
	ctx := context.Background()
	if s.cache != nil {
		if cars, hit, err := s.cache.GetAll(ctx); err != nil {
			s.log.Warn("cars cache get all failed", "error", err)
		} else if hit {
			return cars, nil
		}
	}
	cars, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if errSet := s.cache.SetAll(ctx, cars); errSet != nil {
			s.log.Warn("cars cache set all failed", "error", errSet)
		}
	}
	return cars, nil
}

func (s *CarService) GetByID(id int) (*models.Car, error) {
	ctx := context.Background()
	if s.cache != nil {
		if car, hit, err := s.cache.GetByID(ctx, id); err != nil {
			s.log.Warn("cars cache get by id failed", "error", err, "id", id)
		} else if hit {
			return car, nil
		}
	}
	car, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if car != nil && s.cache != nil {
		if errSet := s.cache.SetByID(ctx, id, car); errSet != nil {
			s.log.Warn("cars cache set by id failed", "error", errSet, "id", id)
		}
	}
	return car, nil
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
	ctx := context.Background()
	if s.cache != nil {
		if cars, hit, err := s.cache.GetFiltered(ctx, name, carType, capacity, priceUnder); err != nil {
			s.log.Warn("cars cache get filtered failed", "error", err)
		} else if hit {
			return cars, nil
		}
	}
	cars, err := s.repo.GetFiltered(name, carType, capacity, priceUnder)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if errSet := s.cache.SetFiltered(ctx, name, carType, capacity, priceUnder, cars); errSet != nil {
			s.log.Warn("cars cache set filtered failed", "error", errSet)
		}
	}
	return cars, nil
}

func (s *CarService) Create(car *models.Car) error {
	if err := s.repo.CreateCar(car); err != nil {
		return err
	}
	s.invalidateCarsCache()
	return nil
}

func (s *CarService) Update(car *models.Car) error {
	if car.ID == 0 {
		return errors.New("car ID must be provided")
	}
	if err := s.repo.UpdateCar(car); err != nil {
		return err
	}
	s.invalidateCarsCache()
	return nil
}

func (s *CarService) Delete(id uint) error {
	if err := s.repo.DeleteCar(id); err != nil {
		return err
	}
	s.invalidateCarsCache()
	return nil
}

func (s *CarService) invalidateCarsCache() {
	if s.cache == nil {
		return
	}
	ctx := context.Background()
	if err := s.cache.InvalidateCars(ctx); err != nil {
		s.log.Warn("cars cache invalidate failed", "error", err)
	}
}
