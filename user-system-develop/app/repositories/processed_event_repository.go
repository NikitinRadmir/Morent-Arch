package repositories

import (
	"user-system/app/models"

	"gorm.io/gorm"
)

type ProcessedEventRepository interface {
	Exists(eventID string) (bool, error)
	Mark(eventID, eventType string) error
}

type processedEventRepository struct {
	db *gorm.DB
}

func NewProcessedEventRepository(db *gorm.DB) ProcessedEventRepository {
	return &processedEventRepository{db: db}
}

func (r *processedEventRepository) Exists(eventID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProcessedEvent{}).Where("event_id = ?", eventID).Count(&count).Error
	return count > 0, err
}

func (r *processedEventRepository) Mark(eventID, eventType string) error {
	return r.db.Create(&models.ProcessedEvent{
		EventID:   eventID,
		EventType: eventType,
	}).Error
}
