package repository

import (
	"gorm.io/gorm"
	"morent-backend/internal/models"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) ListAll() ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Order("created_at DESC").Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) GetByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) GetByCarID(carID int) ([]models.Comment, error) {
	var comments []models.Comment
	errFind := r.db.Where("car_id = ?", carID).Order("created_at DESC").Find(&comments).Error
	return comments, errFind
}

func (r *CommentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) UpdateComment(comment *models.Comment) error {
	result := r.db.Model(&models.Comment{}).Where("id = ?", comment.ID).Updates(comment)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *CommentRepository) DeleteComment(id uint) error {
	result := r.db.Delete(&models.Comment{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// for admin 
func (r *CommentRepository) Update(c *models.Comment) error {
	return r.UpdateComment(c)
}

func (r *CommentRepository) Delete(id uint) error {
	return r.DeleteComment(id)
}
