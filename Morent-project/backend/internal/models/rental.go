package models

import (
	"time"

	"gorm.io/gorm"
)

type Rental struct {
	gorm.Model
	UserID     uint      `gorm:"not null;index"`
	CarID      uint      `gorm:"not null"`
	StartDate  time.Time `gorm:"not null"`
	EndDate    time.Time `gorm:"not null"`
	TotalPrice              float64   `gorm:"type:numeric(10,2);not null"`
	RentalDayReminderSent   bool      `gorm:"not null;default:false"`
	Car                     Car       `gorm:"constraint:OnDelete:CASCADE"`
}

type RentalResponse struct {
	ID         uint      `json:"id"`
	Car        Car       `json:"car"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	TotalPrice float64   `json:"totalPrice"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (r *Rental) ToResponse() RentalResponse {
	return RentalResponse{
		ID:         r.ID,
		Car:        r.Car,
		StartDate:  r.StartDate,
		EndDate:    r.EndDate,
		TotalPrice: r.TotalPrice,
		CreatedAt:  r.CreatedAt,
	}
}
