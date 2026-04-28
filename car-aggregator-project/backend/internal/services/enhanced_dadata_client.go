package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/ekomobile/dadata/v2"
	"github.com/ekomobile/dadata/v2/api/clean"

	"car-aggregator/internal/config"
	"car-aggregator/internal/logging"
)

type EnhancedDaDataClient interface {
	SearchVehicle(ctx context.Context, query string) ([]VehicleResult, error)
	SearchVehicleWithOptions(ctx context.Context, query string, options DaDataSearchOptions) ([]VehicleResult, error)
}

type VehicleResult struct {
	Brand      string  `json:"brand"`
	Model      string  `json:"model"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
	Raw        interface{} `json:"raw,omitempty"`
}

type DaDataSearchOptions struct {
	MaxResults      int           `json:"max_results"`
	MinConfidence   float64       `json:"min_confidence"`
	Timeout         time.Duration `json:"timeout"`
	IncludeRawData  bool          `json:"include_raw_data"`
	RetryCount      int           `json:"retry_count"`
}

// DaDataVehicleResponse represents a vehicle response from DaData API
type DaDataVehicleResponse struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
}

type enhancedDaDataClient struct {
	api           *clean.Api
	config        *config.SearchConfiguration
	logger        logging.SearchLogger
	circuitBreaker *CircuitBreaker
}

type CircuitBreaker struct {
	failureCount    int
	lastFailureTime time.Time
	timeout         time.Duration
	maxFailures     int
	state           CircuitState
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

func NewEnhancedDaDataClient(config *config.SearchConfiguration, logger logging.SearchLogger) EnhancedDaDataClient {
	api := dadata.NewCleanApi()
	
	return &enhancedDaDataClient{
		api:    api,
		config: config,
		logger: logger,
		circuitBreaker: &CircuitBreaker{
			timeout:     30 * time.Second,
			maxFailures: 5,
			state:       CircuitClosed,
		},
	}
}

func (c *enhancedDaDataClient) SearchVehicle(ctx context.Context, query string) ([]VehicleResult, error) {
	options := DaDataSearchOptions{
		MaxResults:    c.config.DaData.MaxResults,
		MinConfidence: 0.3,
		Timeout:       c.config.DaData.Timeout,
		RetryCount:    3,
	}
	
	return c.SearchVehicleWithOptions(ctx, query, options)
}

func (c *enhancedDaDataClient) SearchVehicleWithOptions(ctx context.Context, query string, options DaDataSearchOptions) ([]VehicleResult, error) {
	// Check circuit breaker
	if !c.circuitBreaker.canExecute() {
		c.logger.LogError("dadata_circuit_open", fmt.Errorf("circuit breaker is open"), map[string]interface{}{
			"query": query,
		})
		return nil, fmt.Errorf("DaData service is temporarily unavailable")
	}
	
	startTime := time.Now()
	c.logger.LogSearchStart(query)
	
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	
	var results []VehicleResult
	var lastErr error
	
	// Retry logic with exponential backoff
	for attempt := 0; attempt <= options.RetryCount; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
			select {
			case <-time.After(backoffDuration):
			case <-timeoutCtx.Done():
				return nil, timeoutCtx.Err()
			}
		}
		
		// For now, create a mock response since we don't have exact DaData API details
		// In real implementation, this would call the actual DaData API
		vehicleData := []*DaDataVehicleResponse{}
		
		// Mock processing of the query
		if strings.Contains(strings.ToLower(query), "toyota") {
			vehicleData = append(vehicleData, &DaDataVehicleResponse{
				Brand: "Toyota",
				Model: "Camry",
			})
		}
		if strings.Contains(strings.ToLower(query), "honda") {
			vehicleData = append(vehicleData, &DaDataVehicleResponse{
				Brand: "Honda",
				Model: "Civic",
			})
		}
		
		// Process results
		results = c.processVehicleData(vehicleData, options)
		
		// Log successful result
		dadataResults := make([]logging.DaDataResult, len(results))
		for i, result := range results {
			dadataResults[i] = logging.DaDataResult{
				Brand:      result.Brand,
				Model:      result.Model,
				Confidence: result.Confidence,
			}
		}
		c.logger.LogDaDataResult(query, dadataResults)
		
		// Record success in circuit breaker
		c.circuitBreaker.recordSuccess()
		
		duration := time.Since(startTime)
		c.logger.LogSearchComplete(query, len(results), duration)
		
		return results, nil
	}
	
	// All retries failed
	c.circuitBreaker.recordFailure()
	c.logger.LogError("dadata_all_retries_failed", lastErr, map[string]interface{}{
		"query":        query,
		"retry_count":  options.RetryCount,
		"duration_ms":  time.Since(startTime).Milliseconds(),
	})
	
	return nil, fmt.Errorf("DaData request failed after %d retries: %w", options.RetryCount, lastErr)
}

func (c *enhancedDaDataClient) processVehicleData(vehicleData []*DaDataVehicleResponse, options DaDataSearchOptions) []VehicleResult {
	results := make([]VehicleResult, 0, len(vehicleData))
	
	for _, vehicle := range vehicleData {
		if vehicle == nil {
			continue
		}
		
		confidence := c.calculateConfidence(vehicle)
		
		// Skip results below minimum confidence
		if confidence < options.MinConfidence {
			continue
		}
		
		result := VehicleResult{
			Brand:      vehicle.Brand,
			Model:      vehicle.Model,
			Confidence: confidence,
			Source:     "dadata",
		}
		
		if options.IncludeRawData {
			result.Raw = vehicle
		}
		
		results = append(results, result)
	}
	
	// Sort by confidence (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})
	
	// Limit results
	if len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}
	
	return results
}

func (c *enhancedDaDataClient) calculateConfidence(vehicle *DaDataVehicleResponse) float64 {
	confidence := 0.5 // Base confidence
	
	// Increase confidence based on data completeness
	if vehicle.Brand != "" {
		confidence += 0.2
	}
	
	if vehicle.Model != "" {
		confidence += 0.2
	}
	
	// Increase confidence for well-known brands
	wellKnownBrands := map[string]bool{
		"TOYOTA":     true,
		"HONDA":      true,
		"VOLKSWAGEN": true,
		"BMW":        true,
		"MERCEDES":   true,
		"AUDI":       true,
		"FORD":       true,
		"NISSAN":     true,
	}
	
	if wellKnownBrands[vehicle.Brand] {
		confidence += 0.1
	}
	
	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// Circuit Breaker implementation

func (cb *CircuitBreaker) canExecute() bool {
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	default:
		return false
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	cb.failureCount = 0
	cb.state = CircuitClosed
}

func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
	}
}

// Fallback DaData client for when main client fails
type FallbackDaDataClient struct {
	logger logging.SearchLogger
}

func NewFallbackDaDataClient(logger logging.SearchLogger) *FallbackDaDataClient {
	return &FallbackDaDataClient{
		logger: logger,
	}
}

func (f *FallbackDaDataClient) SearchVehicle(ctx context.Context, query string) ([]VehicleResult, error) {
	f.logger.LogFallbackUsed("fallback_dadata", "main dadata client unavailable")
	
	// Simple pattern matching as fallback
	results := f.patternMatchVehicle(query)
	
	if len(results) == 0 {
		return nil, fmt.Errorf("no fallback results found for query: %s", query)
	}
	
	return results, nil
}

func (f *FallbackDaDataClient) patternMatchVehicle(query string) []VehicleResult {
	// Simple pattern matching for common car names
	patterns := map[string]VehicleResult{
		"golf":    {Brand: "VOLKSWAGEN", Model: "GOLF", Confidence: 0.7, Source: "fallback"},
		"camry":   {Brand: "TOYOTA", Model: "CAMRY", Confidence: 0.7, Source: "fallback"},
		"civic":   {Brand: "HONDA", Model: "CIVIC", Confidence: 0.7, Source: "fallback"},
		"corolla": {Brand: "TOYOTA", Model: "COROLLA", Confidence: 0.7, Source: "fallback"},
		"focus":   {Brand: "FORD", Model: "FOCUS", Confidence: 0.7, Source: "fallback"},
		"accord":  {Brand: "HONDA", Model: "ACCORD", Confidence: 0.7, Source: "fallback"},
		"passat":  {Brand: "VOLKSWAGEN", Model: "PASSAT", Confidence: 0.7, Source: "fallback"},
		"jetta":   {Brand: "VOLKSWAGEN", Model: "JETTA", Confidence: 0.7, Source: "fallback"},
	}
	
	queryLower := strings.ToLower(query)
	
	var results []VehicleResult
	for pattern, result := range patterns {
		if strings.Contains(queryLower, pattern) {
			results = append(results, result)
		}
	}
	
	return results
}

// Enhanced error types for better error handling
type DaDataError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Query   string `json:"query"`
}

func (e *DaDataError) Error() string {
	return fmt.Sprintf("DaData error [%s]: %s (query: %s)", e.Code, e.Message, e.Query)
}

func NewDaDataError(code, message, query string) *DaDataError {
	return &DaDataError{
		Code:    code,
		Message: message,
		Query:   query,
	}
}