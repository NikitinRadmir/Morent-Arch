package interfaces

import (
	"context"
	"time"

	"car-aggregator/internal/dtos"
)

// EnhancedDaDataClient interface for DaData API integration
type EnhancedDaDataClient interface {
	SearchVehicle(ctx context.Context, query string) ([]VehicleResult, error)
	SearchVehicleWithOptions(ctx context.Context, query string, options SearchOptions) ([]VehicleResult, error)
}

// EnhancedCarAPIClient interface for CarAPI integration
type EnhancedCarAPIClient interface {
	Login(ctx context.Context) error
	GetTrimsByMakeAndModel(ctx context.Context, make, model string, limit int) (*dtos.TrimsResponse, error)
	GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error)
	SearchTrims(ctx context.Context, searchTerms []string, limit int) (*dtos.TrimsResponse, error)
	GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error)
}

// VehicleResult represents a vehicle search result
type VehicleResult struct {
	Brand      string  `json:"brand"`
	Model      string  `json:"model"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Raw        interface{} `json:"raw,omitempty"`
}

// SearchOptions provides configuration for search operations
type SearchOptions struct {
	MaxResults      int           `json:"max_results"`
	MinConfidence   float64       `json:"min_confidence"`
	Timeout         time.Duration `json:"timeout"`
	IncludeRawData  bool          `json:"include_raw_data"`
	RetryCount      int           `json:"retry_count"`
}