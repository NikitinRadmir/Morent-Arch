package repository

import (
	"morent-backend/internal/models"

	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(userID uint, token string) error {
	session := models.Session{
		UserID: userID,
		Token:  token,
	}
	return r.db.Create(&session).Error
}

func (r *SessionRepository) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	getSessionErr := r.db.Preload("User").Where("token = ?", token).First(&session).Error
	if getSessionErr != nil {
		if getSessionErr == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, getSessionErr
	}
	return &session, nil
}

func (r *SessionRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&models.Session{}).Error
}
