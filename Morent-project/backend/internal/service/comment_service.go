package service

import (
	"errors"

	"morent-backend/internal/models"
)

type CommentRepository interface {
	GetByCarID(carID int) ([]models.Comment, error)
	GetByID(id uint) (*models.Comment, error)
	CreateComment(comment *models.Comment) error
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uint) error
}

type CommentRentalRepository interface {
	HasRental(userID, carID uint) (bool, error)
}

type CommentService struct {
	repo       CommentRepository
	rentalRepo CommentRentalRepository
}

func NewCommentService(repo CommentRepository, rentalRepo CommentRentalRepository) *CommentService {
	return &CommentService{repo: repo, rentalRepo: rentalRepo}
}

func (s *CommentService) GetByCarID(carID int) ([]models.Comment, error) {
	return s.repo.GetByCarID(carID)
}

func (s *CommentService) GetByID(id uint) (*models.Comment, error) {
	return s.repo.GetByID(id)
}

func (s *CommentService) UserCanComment(userID, carID uint) (bool, error) {
	return s.rentalRepo.HasRental(userID, carID)
}

func (s *CommentService) Create(comment *models.Comment) error {
	return s.repo.CreateComment(comment)
}

func (s *CommentService) Update(comment *models.Comment) error {
	if comment.ID == 0 {
		return errors.New("comment ID must be provided")
	}
	return s.repo.UpdateComment(comment)
}

func (s *CommentService) Delete(id uint) error {
	return s.repo.DeleteComment(id)
}


