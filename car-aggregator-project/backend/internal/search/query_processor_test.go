package search

import (
	"testing"
)

func TestQueryProcessor_ParseQuery(t *testing.T) {
	qp := NewQueryProcessor()
	
	tests := []struct {
		name           string
		query          string
		expectedMake   string
		expectedModel  string
		expectModelSet bool
		expectedYear   *int
		minConfidence  float64
	}{
		{
			name:          "English make and model",
			query:         "Volkswagen Golf",
			expectedMake:  "Volkswagen",
			expectedModel: "Golf",
			minConfidence: 0.8,
		},
		{
			name:          "Russian model name",
			query:         "Гольф",
			expectedMake:  "",
			expectModelSet: true,
			minConfidence: 0.3,
		},
		{
			name:          "With year",
			query:         "2015 Toyota Camry",
			expectedMake:  "Toyota",
			expectedModel: "Camry",
			minConfidence: 0.7,
		},
		{
			name:          "Model only",
			query:         "Civic",
			expectedMake:  "",
			expectedModel: "Civic",
			minConfidence: 0.5,
		},
		{
			name:          "BMW with series",
			query:         "BMW X5",
			expectedMake:  "BMW",
			expectedModel: "X5",
			minConfidence: 0.7,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qp.ParseQuery(tt.query)
			if err != nil {
				t.Errorf("ParseQuery() error = %v", err)
				return
			}
			
			if result.Make != tt.expectedMake {
				t.Errorf("ParseQuery() make = %v, want %v", result.Make, tt.expectedMake)
			}

			if tt.expectModelSet {
				if result.Model == "" {
					t.Errorf("ParseQuery() model is empty, want non-empty")
				}
			} else if result.Model != tt.expectedModel {
				t.Errorf("ParseQuery() model = %v, want %v", result.Model, tt.expectedModel)
			}
			
			if result.Confidence < tt.minConfidence {
				t.Errorf("ParseQuery() confidence = %v, want >= %v", result.Confidence, tt.minConfidence)
			}
			
			if tt.expectedYear != nil {
				if result.Year == nil || *result.Year != *tt.expectedYear {
					t.Errorf("ParseQuery() year = %v, want %v", result.Year, tt.expectedYear)
				}
			}
		})
	}
}

func TestQueryProcessor_NormalizeQuery(t *testing.T) {
	qp := NewQueryProcessor()
	
	tests := []struct {
		name               string
		query              string
		expectedNormalized string
		expectedLanguage   string
		expectNonEmpty     bool
	}{
		{
			name:               "English query",
			query:              "volkswagen golf",
			expectedNormalized: "Volkswagen Golf",
			expectedLanguage:   "en",
		},
		{
			name:               "Russian query",
			query:              "фольксваген гольф",
			expectedLanguage:   "ru",
			expectNonEmpty:     true,
		},
		{
			name:               "Mixed case",
			query:              "TOYOTA camry",
			expectedNormalized: "Toyota Camry",
			expectedLanguage:   "en",
		},
		{
			name:               "Extra spaces",
			query:              "  BMW   X5  ",
			expectedNormalized: "Bmw X5",
			expectedLanguage:   "en",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qp.NormalizeQuery(tt.query)
			if err != nil {
				t.Errorf("NormalizeQuery() error = %v", err)
				return
			}
			
			if tt.expectNonEmpty {
				if result.Normalized == "" {
					t.Errorf("NormalizeQuery() normalized is empty, want non-empty")
				}
			} else if result.Normalized != tt.expectedNormalized {
				t.Errorf("NormalizeQuery() normalized = %v, want %v", result.Normalized, tt.expectedNormalized)
			}
			
			if result.Language != tt.expectedLanguage {
				t.Errorf("NormalizeQuery() language = %v, want %v", result.Language, tt.expectedLanguage)
			}
		})
	}
}

func TestQueryProcessor_ExtractMakeModel(t *testing.T) {
	qp := NewQueryProcessor()
	
	tests := []struct {
		name          string
		query         string
		expectedMake  string
		expectedModel string
		minConfidence float64
	}{
		{
			name:          "Clear make and model",
			query:         "Toyota Camry",
			expectedMake:  "Toyota",
			expectedModel: "Camry",
			minConfidence: 0.7,
		},
		{
			name:          "Model only",
			query:         "Golf",
			expectedMake:  "",
			expectedModel: "Golf",
			minConfidence: 0.5,
		},
		{
			name:          "With year prefix",
			query:         "2020 Honda Civic",
			expectedMake:  "Honda",
			expectedModel: "Civic",
			minConfidence: 0.7,
		},
		{
			name:          "Abbreviation",
			query:         "VW Golf",
			expectedMake:  "Volkswagen",
			expectedModel: "Golf",
			minConfidence: 0.8,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			make, model, confidence := qp.ExtractMakeModel(tt.query)
			
			if make != tt.expectedMake {
				t.Errorf("ExtractMakeModel() make = %v, want %v", make, tt.expectedMake)
			}
			
			if model != tt.expectedModel {
				t.Errorf("ExtractMakeModel() model = %v, want %v", model, tt.expectedModel)
			}
			
			if confidence < tt.minConfidence {
				t.Errorf("ExtractMakeModel() confidence = %v, want >= %v", confidence, tt.minConfidence)
			}
		})
	}
}

func TestQueryProcessor_DetectLanguage(t *testing.T) {
	qp := &queryProcessor{}
	
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "English text",
			query:    "Volkswagen Golf",
			expected: "en",
		},
		{
			name:     "Russian text",
			query:    "Фольксваген Гольф",
			expected: "ru",
		},
		{
			name:     "Mixed with numbers",
			query:    "2015 BMW X5",
			expected: "en",
		},
		{
			name:     "Russian with numbers",
			query:    "2015 БМВ Х5",
			expected: "ru",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qp.detectLanguage(tt.query)
			if result != tt.expected {
				t.Errorf("detectLanguage() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestQueryProcessor_FuzzyMatch(t *testing.T) {
	qp := &queryProcessor{}
	
	tests := []struct {
		name        string
		s1          string
		s2          string
		minSimilarity float64
	}{
		{
			name:        "Exact match",
			s1:          "golf",
			s2:          "golf",
			minSimilarity: 1.0,
		},
		{
			name:        "Close match",
			s1:          "golf",
			s2:          "golff",
			minSimilarity: 0.8,
		},
		{
			name:        "Different words",
			s1:          "golf",
			s2:          "civic",
			minSimilarity: 0.0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qp.fuzzyMatch(tt.s1, tt.s2)
			if tt.minSimilarity == 1.0 && result != 1.0 {
				t.Errorf("fuzzyMatch() = %v, want exactly %v", result, tt.minSimilarity)
			} else if tt.minSimilarity < 1.0 && result < tt.minSimilarity {
				t.Errorf("fuzzyMatch() = %v, want >= %v", result, tt.minSimilarity)
			}
		})
	}
}