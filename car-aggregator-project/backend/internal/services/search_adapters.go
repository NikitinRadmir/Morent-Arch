package services

import (
	"context"
	"car-aggregator/internal/search"
)

// ConvertVehicleResults converts services.VehicleResult to search.VehicleResult
func ConvertVehicleResults(results []VehicleResult) []search.VehicleResult {
	converted := make([]search.VehicleResult, len(results))
	for i, result := range results {
		converted[i] = search.VehicleResult{
			Brand:      result.Brand,
			Model:      result.Model,
			Confidence: result.Confidence,
			Source:     result.Source,
			Raw:        result.Raw,
		}
	}
	return converted
}

// DaDataClientAdapter adapts EnhancedDaDataClient to search.DaDataClient
type DaDataClientAdapter struct {
	client EnhancedDaDataClient
}

func NewDaDataClientAdapter(client EnhancedDaDataClient) *DaDataClientAdapter {
	return &DaDataClientAdapter{client: client}
}

func (adapter *DaDataClientAdapter) SearchVehicle(ctx context.Context, query string) ([]search.VehicleResult, error) {
	results, err := adapter.client.SearchVehicle(ctx, query)
	if err != nil {
		return nil, err
	}
	return ConvertVehicleResults(results), nil
}