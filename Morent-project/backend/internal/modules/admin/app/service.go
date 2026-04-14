package app

import (
	"context"
	"time"

	"morent-backend/internal/models"
	"morent-backend/internal/service"
)

type UserRepository interface {
	ListAll() ([]models.User, error)
	Update(user *models.User) error
	Delete(id uint64) error
}

type RentalRepository interface {
	ListAll() ([]models.Rental, error)
	Delete(id uint) error
}

type FavoriteRepository interface {
	ListAll() ([]models.Favorite, error)
	Remove(userID, carID uint) error
}

type CommentRepository interface {
	ListAll() ([]models.Comment, error)
	Update(c *models.Comment) error
	Delete(id uint) error
}

type LogReader interface {
	GetDailyEvents(ctx context.Context, day time.Time) ([]service.LogEvent, error)
}

type Service struct {
	userRepo     UserRepository
	rentalRepo   RentalRepository
	favoriteRepo FavoriteRepository
	commentRepo  CommentRepository
	logReader    LogReader
}

func NewService(
	userRepo UserRepository,
	rentalRepo RentalRepository,
	favoriteRepo FavoriteRepository,
	commentRepo CommentRepository,
	logReader LogReader,
) *Service {
	return &Service{
		userRepo:     userRepo,
		rentalRepo:   rentalRepo,
		favoriteRepo: favoriteRepo,
		commentRepo:  commentRepo,
		logReader:    logReader,
	}
}

func (s *Service) ListUsers() ([]models.User, error) {
	return s.userRepo.ListAll()
}

func (s *Service) UpdateUser(user *models.User) error {
	return s.userRepo.Update(user)
}

func (s *Service) DeleteUser(id uint64) error {
	return s.userRepo.Delete(id)
}

func (s *Service) ListRentals() ([]models.Rental, error) {
	return s.rentalRepo.ListAll()
}

func (s *Service) DeleteRental(id uint) error {
	return s.rentalRepo.Delete(id)
}

func (s *Service) ListFavorites() ([]models.Favorite, error) {
	return s.favoriteRepo.ListAll()
}

func (s *Service) DeleteFavorite(userID, carID uint) error {
	return s.favoriteRepo.Remove(userID, carID)
}

func (s *Service) ListComments() ([]models.Comment, error) {
	return s.commentRepo.ListAll()
}

func (s *Service) UpdateComment(comment *models.Comment) error {
	return s.commentRepo.Update(comment)
}

func (s *Service) DeleteComment(id uint) error {
	return s.commentRepo.Delete(id)
}

func (s *Service) ListLogs(ctx context.Context, day time.Time) ([]service.LogEvent, error) {
	return s.logReader.GetDailyEvents(ctx, day)
}
