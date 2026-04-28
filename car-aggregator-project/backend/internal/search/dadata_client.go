package search

import (
	"context"
)

// VehicleResult represents a vehicle search result from DaData
type VehicleResult struct {
	Brand      string  `json:"brand"`
	Model      string  `json:"model"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Raw        interface{} `json:"raw,omitempty"`
}

// DaDataClient defines the interface for DaData operations needed by search strategies
type DaDataClient interface {
	SearchVehicle(ctx context.Context, query string) ([]VehicleResult, error)
}