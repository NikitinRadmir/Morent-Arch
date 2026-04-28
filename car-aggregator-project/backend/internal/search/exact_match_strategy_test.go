package search

import (
	"context"
	"fmt"
	"testing"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// Mock CarAPI client for testing
type mockCarAPIClient struct {
	shouldFail bool
	response   *dtos.TrimsResponse
}

func (m *mockCarAPIClient) Login(ctx context.Context) error {
	if m.shouldFail {
		return fmt.Errorf("mock login failed")
	}
	return nil
}

func (m *mockCarAPIClient) GetTrimsByMakeAndModel(ctx context.Context, make, model string, limit int) (*dtos.TrimsResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock API failed")
	}
	return m.response, nil
}

func (m *mockCarAPIClient) GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock API failed")
	}
	return m.response, nil
}

func (m *mockCarAPIClient) SearchTrims(ctx context.Context, searchTerms []string, limit int) (*dtos.TrimsResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock API failed")
	}
	return m.response, nil
}

func (m *mockCarAPIClient) GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error) {
	if m.shouldFail {
		return 0, 0, fmt.Errorf("mock specs failed")
	}
	return 5, 7.5, nil // Default specs
}

func TestExactMatchStrategy_CanHandle(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	mockClient := &mockCarAPIClient{}
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	tests := []struct {
		name     string
		query    *ParsedQuery
		expected bool
	}{
		{
			name: "High confidence with make and model",
			query: &ParsedQuery{
				Make:       "Toyota",
				Model:      "Camry",
				Confidence: 0.9,
			},
			expected: true,
		},
		{
			name: "Low confidence",
			query: &ParsedQuery{
				Make:       "Toyota",
				Model:      "Camry",
				Confidence: 0.5,
			},
			expected: false,
		},
		{
			name: "Missing make",
			query: &ParsedQuery{
				Model:      "Camry",
				Confidence: 0.9,
			},
			expected: false,
		},
		{
			name: "Missing model",
			query: &ParsedQuery{
				Make:       "Toyota",
				Confidence: 0.9,
			},
			expected: false,
		},
		{
			name: "Empty make and model",
			query: &ParsedQuery{
				Confidence: 0.9,
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.CanHandle(tt.query)
			if result != tt.expected {
				t.Errorf("CanHandle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExactMatchStrategy_Search(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	
	// Create mock response
	mockResponse := &dtos.TrimsResponse{
		Data: []dtos.Trim{
			{
				ID:          1,
				Year:        2020,
				Make:        "Toyota",
				Model:       "Camry",
				Trim:        "LE",
				Description: "LE 4dr Sedan (2.5L 4cyl 8A)",
				MSRP:        24970,
			},
			{
				ID:          2,
				Year:        2021,
				Make:        "Toyota",
				Model:       "Camry",
				Trim:        "XLE",
				Description: "XLE 4dr Sedan (2.5L 4cyl 8A)",
				MSRP:        28360,
			},
		},
	}
	
	mockClient := &mockCarAPIClient{
		shouldFail: false,
		response:   mockResponse,
	}
	
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	query := &ParsedQuery{
		Original:   "Toyota Camry",
		Make:       "Toyota",
		Model:      "Camry",
		Confidence: 0.9,
	}
	
	result, err := strategy.Search(context.Background(), query)
	if err != nil {
		t.Errorf("Search() error = %v", err)
		return
	}
	
	if result == nil {
		t.Error("Search() returned nil result")
		return
	}
	
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
	
	if result.Source != StrategyNameExactMatch {
		t.Errorf("Expected source %s, got %s", StrategyNameExactMatch, result.Source)
	}
	
	if result.Confidence <= 0.8 {
		t.Errorf("Expected high confidence (>0.8), got %f", result.Confidence)
	}
	
	// Check first item
	if len(result.Items) > 0 {
		item := result.Items[0]
		if item.Make != "Toyota" || item.Model != "Camry" {
			t.Errorf("Expected Toyota Camry, got %s %s", item.Make, item.Model)
		}
	}
}

func TestExactMatchStrategy_SearchWithYear(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	
	mockResponse := &dtos.TrimsResponse{
		Data: []dtos.Trim{
			{
				ID:          1,
				Year:        2020,
				Make:        "Toyota",
				Model:       "Camry",
				Trim:        "LE",
				Description: "LE 4dr Sedan (2.5L 4cyl 8A)",
				MSRP:        24970,
			},
			{
				ID:          2,
				Year:        2021,
				Make:        "Toyota",
				Model:       "Camry",
				Trim:        "XLE",
				Description: "XLE 4dr Sedan (2.5L 4cyl 8A)",
				MSRP:        28360,
			},
		},
	}
	
	mockClient := &mockCarAPIClient{
		shouldFail: false,
		response:   mockResponse,
	}
	
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	year := 2020
	query := &ParsedQuery{
		Original:   "2020 Toyota Camry",
		Make:       "Toyota",
		Model:      "Camry",
		Year:       &year,
		Confidence: 0.9,
	}
	
	result, err := strategy.Search(context.Background(), query)
	if err != nil {
		t.Errorf("Search() error = %v", err)
		return
	}
	
	if result == nil {
		t.Error("Search() returned nil result")
		return
	}
	
	// Should only return 2020 models (year filter applied)
	if len(result.Items) != 1 {
		t.Errorf("Expected 1 item (2020 only), got %d", len(result.Items))
	}
	
	if len(result.Items) > 0 && result.Items[0].Year != 2020 {
		t.Errorf("Expected year 2020, got %d", result.Items[0].Year)
	}
}

func TestExactMatchStrategy_SearchFailure(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	
	mockClient := &mockCarAPIClient{
		shouldFail: true,
	}
	
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	query := &ParsedQuery{
		Original:   "Toyota Camry",
		Make:       "Toyota",
		Model:      "Camry",
		Confidence: 0.9,
	}
	
	result, err := strategy.Search(context.Background(), query)
	if err == nil {
		t.Error("Search() should have returned an error")
	}
	
	if result != nil {
		t.Error("Search() should have returned nil result on error")
	}
}

func TestExactMatchStrategy_NormalizeMake(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	mockClient := &mockCarAPIClient{}
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	tests := []struct {
		input    string
		expected string
	}{
		{"vw", "Volkswagen"},
		{"VW", "Volkswagen"},
		{"volkswagen", "Volkswagen"},
		{"bmw", "BMW"},
		{"BMW", "BMW"},
		{"mercedes", "Mercedes-Benz"},
		{"benz", "Mercedes-Benz"},
		{"toyota", "Toyota"},
		{"TOYOTA", "Toyota"},
		{"unknown", "Unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := strategy.normalizeMake(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeMake(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExactMatchStrategy_NormalizeModel(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	mockClient := &mockCarAPIClient{}
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	tests := []struct {
		input    string
		expected string
	}{
		{"golf", "Golf"},
		{"GOLF", "Golf"},
		{"camry", "Camry"},
		{"civic", "Civic"},
		{"x5", "X5"},
		{"X5", "X5"},
		{"a4", "A4"},
		{"A4", "A4"},
		{"q7", "Q7"},
		{"f-150", "F-150"},
		{"f150", "F-150"},
		{"cr-v", "CR-V"},
		{"crv", "CR-V"},
		{"unknown", "Unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := strategy.normalizeModel(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeModel(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExactMatchStrategy_InferTransmission(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	mockClient := &mockCarAPIClient{}
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	tests := []struct {
		description string
		trim        string
		expected    string
	}{
		{"LE 4dr Sedan (2.5L 4cyl 8A)", "", "Automatic"},
		{"S 2dr Hatchback (1.8L 4cyl Turbo 5M)", "", "Manual"},
		{"CVT transmission", "", "Automatic"},
		{"6-speed manual", "", "Manual"},
		{"Automatic transmission", "", "Automatic"},
		{"Standard transmission", "", "Manual"},
		{"Unknown transmission", "", "Manual"}, // Default
	}
	
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := strategy.inferTransmission(tt.description, tt.trim)
			if result != tt.expected {
				t.Errorf("inferTransmission(%s, %s) = %s, want %s", tt.description, tt.trim, result, tt.expected)
			}
		})
	}
}

func TestExactMatchStrategy_ValidateExactMatch(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	mockClient := &mockCarAPIClient{}
	strategy := NewExactMatchStrategy(mockClient, logger)
	
	query := &ParsedQuery{
		Make:  "Toyota",
		Model: "Camry",
	}
	
	tests := []struct {
		name     string
		items    []dtos.SearchTrimItem
		expected bool
	}{
		{
			name: "Exact match found",
			items: []dtos.SearchTrimItem{
				{Make: "Toyota", Model: "Camry", Year: 2020},
				{Make: "Honda", Model: "Civic", Year: 2021},
			},
			expected: true,
		},
		{
			name: "No exact match",
			items: []dtos.SearchTrimItem{
				{Make: "Honda", Model: "Civic", Year: 2021},
				{Make: "Ford", Model: "Focus", Year: 2020},
			},
			expected: false,
		},
		{
			name:     "Empty items",
			items:    []dtos.SearchTrimItem{},
			expected: false,
		},
		{
			name: "Case insensitive match",
			items: []dtos.SearchTrimItem{
				{Make: "TOYOTA", Model: "CAMRY", Year: 2020},
			},
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strategy.ValidateExactMatch(query, tt.items)
			if result != tt.expected {
				t.Errorf("ValidateExactMatch() = %v, want %v", result, tt.expected)
			}
		})
	}
}