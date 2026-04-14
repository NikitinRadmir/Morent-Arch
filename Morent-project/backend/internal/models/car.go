package models

import (
	"time"
	"gorm.io/gorm"
)

type Car struct {
	ID          uint           `json:"id" gorm:"primaryKey"`           // Уникальный идентификатор
	Name        string         `json:"name" gorm:"not null"`          // Название машины (например, "Skoda Octavia RS")
	Type        string         `json:"type" gorm:"not null"`          // Тип машины (Sport, SUV, Sedan, Electric)
	Capacity    int            `json:"capacity" gorm:"not null"`      // Количество мест
	Price       float64        `json:"price" gorm:"type:decimal(10,2);not null"` // Цена за день в долларах
	Description string         `json:"description"`                   // Описание машины
	ImgSrc      string         `json:"imgSrc" gorm:"column:img_src"` // Путь к изображению
	Fuel        float64        `json:"fuel" gorm:"type:decimal(5,2);default:0"` // Расход топлива в литрах
	Transmission string         `json:"transmission" gorm:"not null"` // Тип трансмиссии (Manual, Automatic)
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	Comments    []Comment      `json:"-" gorm:"foreignKey:CarID"`
}

func (Car) TableName() string {
	return "cars"
}



