package httpdto

import "time"

type RentalRequest struct {
	CarID      uint    `json:"carId" validate:"required"`
	StartDate  string  `json:"startDate" validate:"required"`
	EndDate    string  `json:"endDate" validate:"required"`
	TotalPrice float64 `json:"totalPrice" validate:"gte=0"`
}

type BookingDTO struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}
