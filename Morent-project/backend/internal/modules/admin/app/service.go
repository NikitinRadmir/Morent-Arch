package app

import (
	"context"
	"errors"
	"strings"
	"time"

	admindto "morent-backend/internal/modules/admin/httpdto"
	"morent-backend/internal/models"
	"morent-backend/internal/service"
)

type UserRepository interface {
	ListAll() ([]models.User, error)
	GetByID(id uint) (*models.User, error)
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
	GetByID(id uint) (*models.Comment, error)
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

func (s *Service) GetUser(id uint) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *Service) UpdateUser(req admindto.UpdateUserRequest) error {
	user, err := s.userRepo.GetByID(req.ID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if req.Name != nil {
		user.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		user.Email = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.Nickname != nil {
		user.Nickname = strings.TrimSpace(*req.Nickname)
	}
	if req.Position != nil {
		user.Position = strings.TrimSpace(*req.Position)
	}
	if req.AvatarURL != nil {
		user.AvatarURL = strings.TrimSpace(*req.AvatarURL)
	}
	if req.Role != nil {
		role := strings.ToLower(strings.TrimSpace(*req.Role))
		if role != "admin" && role != "user" {
			return errors.New("invalid role")
		}
		user.Role = role
	}
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

func (s *Service) UpdateComment(req admindto.UpdateCommentRequest) error {
	target, err := s.commentRepo.GetByID(req.ID)
	if err != nil {
		return err
	}
	if target == nil {
		return errors.New("comment not found")
	}
	if req.Description != nil {
		target.Description = strings.TrimSpace(*req.Description)
	}
	if req.Rating != nil {
		target.Rating = *req.Rating
	}
	return s.commentRepo.Update(target)
}

func (s *Service) DeleteComment(id uint) error {
	return s.commentRepo.Delete(id)
}

func (s *Service) ListLogs(ctx context.Context, day time.Time) ([]service.LogEvent, error) {
	return s.logReader.GetDailyEvents(ctx, day)
}
