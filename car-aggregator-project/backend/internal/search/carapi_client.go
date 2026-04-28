package search

import (
	"context"
	"car-aggregator/internal/dtos"
)

// CarAPIClient defines the interface for CarAPI operations needed by search strategies
type CarAPIClient interface {
	GetTrimsByMakeAndModel(ctx context.Context, make, model string, limit int) (*dtos.TrimsResponse, error)
	GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error)
	GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error)
}