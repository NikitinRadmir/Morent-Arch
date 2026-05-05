package service

import (
	"errors"

	"morent-backend/internal/models"
)

var ErrCarNotFound = errors.New("car not found")

type FavoriteRepository interface {
	ListByUser(userID uint) ([]models.Favorite, error)
	Create(userID, carID uint) error
	Delete(userID, carID uint) error
}

type FavoriteCarRepository interface {
	GetByID(id int) (*models.Car, error)
}

type FavoriteService struct {
	favoriteRepo FavoriteRepository
	carRepo      FavoriteCarRepository
}

func NewFavoriteService(favRepo FavoriteRepository, carRepo FavoriteCarRepository) *FavoriteService {
	return &FavoriteService{
		favoriteRepo: favRepo,
		carRepo:      carRepo,
	}
}

func (s *FavoriteService) List(userID uint) ([]models.Car, error) {
	favs, listFavoritesErr := s.favoriteRepo.ListByUser(userID)
	if listFavoritesErr != nil {
		return nil, listFavoritesErr
	}
	cars := make([]models.Car, 0, len(favs))
	for _, fav := range favs {
		if fav.Car.ID != 0 {
			cars = append(cars, fav.Car)
		}
	}
	return cars, nil
}

func (s *FavoriteService) Add(userID, carID uint) error {
	car, getCarByIDErr := s.carRepo.GetByID(int(carID))
	if getCarByIDErr != nil {
		return getCarByIDErr
	}
	if car == nil {
		return ErrCarNotFound
	}
	return s.favoriteRepo.Create(userID, carID)
}

func (s *FavoriteService) Remove(userID, carID uint) error {
	return s.favoriteRepo.Delete(userID, carID)
}
