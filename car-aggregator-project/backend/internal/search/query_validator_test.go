package search

import (
	"testing"
)

func TestQueryValidator_ValidateQuery(t *testing.T) {
	qv := NewQueryValidator()
	
	tests := []struct {
		name      string
		query     string
		wantError bool
	}{
		{
			name:      "Valid query",
			query:     "Toyota Camry",
			wantError: false,
		},
		{
			name:      "Valid Russian query",
			query:     "Фольксваген Гольф",
			wantError: false,
		},
		{
			name:      "Valid query with year",
			query:     "2015 BMW X5",
			wantError: false,
		},
		{
			name:      "Empty query",
			query:     "",
			wantError: true,
		},
		{
			name:      "Too short query",
			query:     " ",
			wantError: true,
		},
		{
			name:      "Query with invalid characters",
			query:     "Toyota<script>alert('xss')</script>",
			wantError: true,
		},
		{
			name:      "SQL injection attempt",
			query:     "'; DROP TABLE cars; --",
			wantError: true,
		},
		{
			name:      "Valid query with hyphen",
			query:     "Mercedes-Benz C-Class",
			wantError: false,
		},
		{
			name:      "Valid query with dot",
			query:     "Audi A4 2.0T",
			wantError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qv.ValidateQuery(tt.query)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateQuery() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestQueryValidator_SanitizeQuery(t *testing.T) {
	qv := NewQueryValidator()
	
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "Normal query",
			query:    "Toyota Camry",
			expected: "Toyota Camry",
		},
		{
			name:     "Query with extra spaces",
			query:    "  Toyota   Camry  ",
			expected: "Toyota Camry",
		},
		{
			name:     "Query with special characters",
			query:    "Toyota@#$%Camry!",
			expected: "ToyotaCamry",
		},
		{
			name:     "Query with multiple dots",
			query:    "Audi A4...2.0T",
			expected: "Audi A4.2.0T",
		},
		{
			name:     "Query with multiple hyphens",
			query:    "Mercedes---Benz",
			expected: "Mercedes-Benz",
		},
		{
			name:     "Russian query with spaces",
			query:    "  Фольксваген  Гольф  ",
			expected: "Фольксваген Гольф",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qv.SanitizeQuery(tt.query)
			if result != tt.expected {
				t.Errorf("SanitizeQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestQueryValidator_PreprocessQuery(t *testing.T) {
	qv := NewQueryValidator()
	
	tests := []struct {
		name         string
		query        string
		expectValid  bool
		expectYear   bool
		expectMake   bool
		expectModel  bool
	}{
		{
			name:        "Complete query with year, make, and model",
			query:       "2015 Toyota Camry",
			expectValid: true,
			expectYear:  true,
			expectMake:  true,
			expectModel: true,
		},
		{
			name:        "Make and model only",
			query:       "Honda Civic",
			expectValid: true,
			expectYear:  false,
			expectMake:  true,
			expectModel: true,
		},
		{
			name:        "Model only",
			query:       "Golf",
			expectValid: true,
			expectYear:  false,
			expectMake:  false,
			expectModel: true,
		},
		{
			name:        "BMW X-series",
			query:       "BMW X5",
			expectValid: true,
			expectYear:  false,
			expectMake:  true,
			expectModel: true,
		},
		{
			name:        "Audi A-series",
			query:       "Audi A4",
			expectValid: true,
			expectYear:  false,
			expectMake:  true,
			expectModel: true,
		},
		{
			name:        "Invalid empty query",
			query:       "",
			expectValid: false,
			expectYear:  false,
			expectMake:  false,
			expectModel: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qv.PreprocessQuery(tt.query)
			
			if tt.expectValid && err != nil {
				t.Errorf("PreprocessQuery() unexpected error = %v", err)
				return
			}
			
			if !tt.expectValid && err == nil {
				t.Errorf("PreprocessQuery() expected error but got none")
				return
			}
			
			if tt.expectValid {
				if result.IsValid != tt.expectValid {
					t.Errorf("PreprocessQuery() IsValid = %v, want %v", result.IsValid, tt.expectValid)
				}
				
				if result.HasYear != tt.expectYear {
					t.Errorf("PreprocessQuery() HasYear = %v, want %v", result.HasYear, tt.expectYear)
				}
				
				if result.HasMake != tt.expectMake {
					t.Errorf("PreprocessQuery() HasMake = %v, want %v", result.HasMake, tt.expectMake)
				}
				
				if result.HasModel != tt.expectModel {
					t.Errorf("PreprocessQuery() HasModel = %v, want %v", result.HasModel, tt.expectModel)
				}
			}
		})
	}
}

func TestQueryValidator_GenerateSuggestions(t *testing.T) {
	qv := &queryValidator{
		knownMakes:  initKnownMakes(),
		knownModels: initKnownModels(),
	}
	
	tests := []struct {
		name            string
		query           string
		expectSuggestions bool
	}{
		{
			name:            "Model only should suggest makes",
			query:           "Golf",
			expectSuggestions: true,
		},
		{
			name:            "Very short query should suggest popular searches",
			query:           "a",
			expectSuggestions: true,
		},
		{
			name:            "Typo should suggest correction",
			query:           "golf",
			expectSuggestions: true,
		},
		{
			name:            "Complete query should have fewer suggestions",
			query:           "Toyota Camry",
			expectSuggestions: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := qv.generateSuggestions(tt.query)
			
			hasSuggestions := len(suggestions) > 0
			if hasSuggestions != tt.expectSuggestions {
				t.Errorf("generateSuggestions() has suggestions = %v, want %v", hasSuggestions, tt.expectSuggestions)
			}
		})
	}
}

func TestIsValidCarQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{
			name:  "Valid car query",
			query: "Toyota",
			want:  true,
		},
		{
			name:  "Valid car query with numbers",
			query: "BMW X5",
			want:  true,
		},
		{
			name:  "Empty query",
			query: "",
			want:  false,
		},
		{
			name:  "Only spaces",
			query: "   ",
			want:  false,
		},
		{
			name:  "Only numbers",
			query: "123",
			want:  false,
		},
		{
			name:  "Single character",
			query: "a",
			want:  false,
		},
		{
			name:  "Valid two characters",
			query: "vw",
			want:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidCarQuery(tt.query); got != tt.want {
				t.Errorf("IsValidCarQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeSpacing(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Normal spacing",
			input: "Toyota Camry",
			want:  "Toyota Camry",
		},
		{
			name:  "Multiple spaces",
			input: "Toyota    Camry",
			want:  "Toyota Camry",
		},
		{
			name:  "Leading and trailing spaces",
			input: "  Toyota Camry  ",
			want:  "Toyota Camry",
		},
		{
			name:  "Mixed spacing issues",
			input: "  Toyota   Camry   2015  ",
			want:  "Toyota Camry 2015",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeSpacing(tt.input); got != tt.want {
				t.Errorf("NormalizeSpacing() = %v, want %v", got, tt.want)
			}
		})
	}
}