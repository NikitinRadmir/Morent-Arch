package models

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID          uint           `json:"id" gorm:"primaryKey"`        // Уникальный идентификатор
	CarID       uint           `json:"carId" gorm:"not null;index"` // ID машины, к которой относится отзыв
	UserID      uint           `json:"userId" gorm:"index"`         // Пользователь, оставивший отзыв (для старых сидов может быть 0)
	Name        string         `json:"name" gorm:"not null"`         // Имя автора отзыва
	Post        string         `json:"post"`                         // Должность/роль автора
	Photo       string         `json:"photo"`                        // Путь к фото автора
	Date        string         `json:"date"`                         // Дата отзыва (например, "2 days ago")
	Rating      int            `json:"rating"`                       // Оценка 1-5
	Description string         `json:"description" gorm:"not null"`  // Текст отзыва
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Car Car `json:"-" gorm:"foreignKey:CarID"`
}

func (Comment) TableName() string {
	return "comments"
}
