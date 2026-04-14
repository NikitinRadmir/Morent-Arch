package httpdto

import "morent-backend/internal/models"

type FilterQuery struct {
	Name      string
	CarType   string
	Capacity  *int
	PriceUnder *float64
}

type UpsertCarRequest struct {
	models.Car
}
