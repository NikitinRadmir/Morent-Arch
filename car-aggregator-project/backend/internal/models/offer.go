package models

import "time"

type VehicleCondition string

const (
	ConditionNew       VehicleCondition = "new"
	ConditionUsed      VehicleCondition = "used"
	ConditionCertified VehicleCondition = "certified"
	ConditionUnknown   VehicleCondition = "unknown"
)

type OfferStatus string

const (
	StatusAvailable OfferStatus = "available"
	StatusReserved  OfferStatus = "reserved"
	StatusSold      OfferStatus = "sold"
)

type Offer struct {
	ID int64 `json:"id" gorm:"primaryKey"`

	// Кто "забрал" авто (например, morent).
	CustomerService string `json:"customer_service" gorm:"size:100;not null;index"`

	// ID машины в источнике (в нашем случае удобнее хранить ID trim из carapi).
	CarID int `json:"car_id" gorm:"not null;uniqueIndex"`

	Make  string `json:"make" gorm:"size:100;not null"`
	Model string `json:"model" gorm:"size:100;not null"`
	Trim  string `json:"trim" gorm:"size:150;not null"`

	Price int `json:"price" gorm:"not null"`

	CreatedAt time.Time `json:"created_at"`
}

// PurchaseRequest представляет запрос на "покупку" автомобиля
type PurchaseRequest struct {
	CustomerService string `json:"service_name" binding:"required"`
	CarID           int    `json:"car_id" binding:"required"`
	Make            string `json:"make" binding:"required"`
	Model           string `json:"model" binding:"required"`
	Trim            string `json:"trim" binding:"required"`
	Price           int    `json:"price" binding:"required"`
}

// ReleaseRequest представляет запрос на "освобождение" автомобиля
type ReleaseRequest struct {
	CustomerService string `json:"service_name" binding:"required"`
	CarID           int    `json:"car_id" binding:"required"`
}
