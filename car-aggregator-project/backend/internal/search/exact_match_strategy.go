package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// ExactMatchStrategy implements precise make+model searches
type ExactMatchStrategy struct {
	*BaseStrategy
	carAPIClient CarAPIClient
}

// NewExactMatchStrategy creates a new exact match strategy
func NewExactMatchStrategy(carAPIClient CarAPIClient, logger logging.SearchLogger) *ExactMatchStrategy {
	return &ExactMatchStrategy{
		BaseStrategy: NewBaseStrategy(StrategyNameExactMatch, PriorityExactMatch, logger),
		carAPIClient: carAPIClient,
	}
}

// CanHandle determines if this strategy can handle the given query
func (ems *ExactMatchStrategy) CanHandle(query *ParsedQuery) bool {
	// Can handle if we have both make and model with high confidence
	return query.Make != "" && query.Model != "" && query.Confidence >= 0.8
}

// Search performs exact match search using make and model
func (ems *ExactMatchStrategy) Search(ctx context.Context, query *ParsedQuery) (*SearchResult, error) {
	ems.LogStart(query)
	startTime := time.Now()
	
	// Try exact match with make and model
	result, err := ems.searchExactMatch(ctx, query)
	
	duration := time.Since(startTime)
	
	if err != nil {
		ems.LogResult(query, nil, err)
		return nil, err
	}
	
	result.Duration = duration
	ems.LogResult(query, result, nil)
	
	return result, nil
}

// searchExactMatch performs the actual exact match search
func (ems *ExactMatchStrategy) searchExactMatch(ctx context.Context, query *ParsedQuery) (*SearchResult, error) {
	// Normalize make and model for better matching
	normalizedMake := ems.normalizeMake(query.Make)
	normalizedModel := ems.normalizeModel(query.Model)
	
	// Try with normalized make and model
	response, err := ems.carAPIClient.GetTrimsByMakeAndModel(ctx, normalizedMake, normalizedModel, 10)
	if err != nil {
		return nil, err
	}
	
	if response == nil || len(response.Data) == 0 {
		// Try with original values if normalized search fails
		response, err = ems.carAPIClient.GetTrimsByMakeAndModel(ctx, query.Make, query.Model, 10)
		if err != nil {
			return nil, err
		}
	}
	
	if response == nil || len(response.Data) == 0 {
		return &SearchResult{
			Items:      []dtos.SearchTrimItem{},
			Source:     ems.Name(),
			Confidence: 0.0,
		}, nil
	}
	
	// Convert CarAPI response to SearchTrimItems
	items := ems.convertToSearchItems(response.Data, query)
	
	// Calculate confidence based on match quality
	confidence := ems.calculateConfidence(query, items)
	
	return &SearchResult{
		Items:      items,
		Source:     ems.Name(),
		Confidence: confidence,
		Metadata: map[string]interface{}{
			"make":         normalizedMake,
			"model":        normalizedModel,
			"total_trims":  len(response.Data),
			"search_type":  "exact_match",
		},
	}, nil
}

// convertToSearchItems converts CarAPI trims to SearchTrimItems
func (ems *ExactMatchStrategy) convertToSearchItems(trims []dtos.Trim, query *ParsedQuery) []dtos.SearchTrimItem {
	items := make([]dtos.SearchTrimItem, 0, len(trims))
	
	for _, trim := range trims {
		// Get additional specs if available
		seats, fuel, _ := ems.carAPIClient.GetVehicleSpecs(context.Background(), trim.Year, trim.Make, trim.Model)
		
		// Infer transmission from description
		transmission := ems.inferTransmission(trim.Description, trim.Trim)
		
		// Generate image URL
		imageURL := ems.buildImageURL(trim.Year, trim.Make, trim.Model, trim.Trim)
		
		item := dtos.SearchTrimItem{
			ID:           trim.ID,
			Year:         trim.Year,
			Make:         trim.Make,
			Model:        trim.Model,
			Trim:         trim.Trim,
			Description:  trim.Description,
			MSRP:         trim.MSRP,
			Transmission: transmission,
			Seats:        seats,
			Fuel:         fuel,
			ImageURL:     imageURL,
		}
		
		// Apply year filter if specified in query
		if query.Year != nil && trim.Year != *query.Year {
			continue
		}
		
		items = append(items, item)
	}
	
	return items
}

// calculateConfidence calculates confidence based on match quality
func (ems *ExactMatchStrategy) calculateConfidence(query *ParsedQuery, items []dtos.SearchTrimItem) float64 {
	if len(items) == 0 {
		return 0.0
	}
	
	confidence := 0.9 // High base confidence for exact match
	
	// Check if make matches exactly
	if len(items) > 0 {
		firstItem := items[0]
		if strings.EqualFold(firstItem.Make, query.Make) {
			confidence += 0.05
		}
		
		if strings.EqualFold(firstItem.Model, query.Model) {
			confidence += 0.05
		}
		
		// Bonus for year match
		if query.Year != nil && firstItem.Year == *query.Year {
			confidence += 0.05
		}
	}
	
	// Adjust based on number of results (more results = higher confidence)
	if len(items) >= 5 {
		confidence += 0.02
	} else if len(items) >= 10 {
		confidence += 0.05
	}
	
	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// normalizeMake normalizes make names for better matching
func (ems *ExactMatchStrategy) normalizeMake(make string) string {
	makeMap := map[string]string{
		"vw":         "Volkswagen",
		"volkswagen": "Volkswagen",
		"bmw":        "BMW",
		"mercedes":   "Mercedes-Benz",
		"benz":       "Mercedes-Benz",
		"mb":         "Mercedes-Benz",
		"audi":       "Audi",
		"toyota":     "Toyota",
		"honda":      "Honda",
		"ford":       "Ford",
		"nissan":     "Nissan",
		"hyundai":    "Hyundai",
		"kia":        "Kia",
		"mazda":      "Mazda",
		"subaru":     "Subaru",
		"chevrolet":  "Chevrolet",
		"chevy":      "Chevrolet",
		"gmc":        "GMC",
		"cadillac":   "Cadillac",
		"lexus":      "Lexus",
		"infiniti":   "Infiniti",
		"acura":      "Acura",
	}
	
	makeLower := strings.ToLower(strings.TrimSpace(make))
	if normalized, exists := makeMap[makeLower]; exists {
		return normalized
	}
	
	// Title case the make if no mapping found
	return strings.Title(makeLower)
}

// normalizeModel normalizes model names for better matching
func (ems *ExactMatchStrategy) normalizeModel(model string) string {
	modelMap := map[string]string{
		"golf":     "Golf",
		"passat":   "Passat",
		"jetta":    "Jetta",
		"tiguan":   "Tiguan",
		"touareg":  "Touareg",
		"beetle":   "Beetle",
		"camry":    "Camry",
		"corolla":  "Corolla",
		"prius":    "Prius",
		"rav4":     "RAV4",
		"highlander": "Highlander",
		"sienna":   "Sienna",
		"tacoma":   "Tacoma",
		"civic":    "Civic",
		"accord":   "Accord",
		"cr-v":     "CR-V",
		"crv":      "CR-V",
		"pilot":    "Pilot",
		"fit":      "Fit",
		"ridgeline": "Ridgeline",
		"odyssey":  "Odyssey",
		"focus":    "Focus",
		"fiesta":   "Fiesta",
		"mondeo":   "Mondeo",
		"kuga":     "Kuga",
		"explorer": "Explorer",
		"f-150":    "F-150",
		"f150":     "F-150",
		"mustang":  "Mustang",
		"sentra":   "Sentra",
		"altima":   "Altima",
		"maxima":   "Maxima",
		"rogue":    "Rogue",
		"pathfinder": "Pathfinder",
		"titan":    "Titan",
		"leaf":     "Leaf",
	}
	
	modelLower := strings.ToLower(strings.TrimSpace(model))
	if normalized, exists := modelMap[modelLower]; exists {
		return normalized
	}
	
	// Handle special cases
	model = strings.TrimSpace(model)
	
	// BMW series normalization
	if strings.Contains(strings.ToLower(model), "series") {
		return strings.Title(strings.ToLower(model))
	}
	
	// Handle X-series (X1, X2, etc.)
	if len(model) == 2 && strings.ToUpper(model[:1]) == "X" {
		return strings.ToUpper(model)
	}
	
	// Handle A/Q series for Audi (A3, A4, Q5, etc.)
	if len(model) == 2 && (strings.ToUpper(model[:1]) == "A" || strings.ToUpper(model[:1]) == "Q") {
		return strings.ToUpper(model)
	}
	
	// Default: title case
	return strings.Title(strings.ToLower(model))
}

// inferTransmission infers transmission type from description and trim
func (ems *ExactMatchStrategy) inferTransmission(description string, trimName string) string {
	combined := strings.ToLower(description + " " + trimName)
	
	// Look for automatic indicators
	automaticIndicators := []string{
		"cvt", "automatic", " auto", " 6a", " 7a", " 8a", " 9a", " 10a",
		"6-speed automatic", "7-speed automatic", "8-speed automatic",
		"continuously variable", "dual-clutch", "dct",
	}
	
	for _, indicator := range automaticIndicators {
		if strings.Contains(combined, indicator) {
			return "Automatic"
		}
	}
	
	// Look for manual indicators
	manualIndicators := []string{
		"manual", " 5m", " 6m", "5-speed manual", "6-speed manual",
		"stick", "mt", "standard transmission",
	}
	
	for _, indicator := range manualIndicators {
		if strings.Contains(combined, indicator) {
			return "Manual"
		}
	}
	
	// Default to Manual if uncertain
	return "Manual"
}

// buildImageURL builds an image URL for the vehicle
func (ems *ExactMatchStrategy) buildImageURL(year int, make, model, trim string) string {
	// Use a simple image service (this is a placeholder implementation)
	// In a real implementation, you might use a proper image service
	query := strings.TrimSpace(fmt.Sprintf("%d %s %s", year, make, model))
	if query == "" {
		return ""
	}
	
	// Simple base64 encoding for the query
	encoded := base64EncodeString(query)
	return "https://www.regcheck.org.uk/image.aspx/@" + encoded
}

// Simple base64 encoding function (placeholder)
func base64EncodeString(s string) string {
	// This is a simplified implementation
	// In a real application, use proper base64 encoding
	return strings.ReplaceAll(s, " ", "_")
}

// ValidateExactMatch validates if the search results match the query exactly
func (ems *ExactMatchStrategy) ValidateExactMatch(query *ParsedQuery, items []dtos.SearchTrimItem) bool {
	if len(items) == 0 {
		return false
	}
	
	// Check if at least one item matches make and model exactly
	for _, item := range items {
		makeMatch := strings.EqualFold(item.Make, query.Make)
		modelMatch := strings.EqualFold(item.Model, query.Model)
		
		if makeMatch && modelMatch {
			return true
		}
	}
	
	return false
}

// GetMatchQuality returns a quality score for the match
func (ems *ExactMatchStrategy) GetMatchQuality(query *ParsedQuery, items []dtos.SearchTrimItem) float64 {
	if len(items) == 0 {
		return 0.0
	}
	
	totalScore := 0.0
	for _, item := range items {
		score := 0.0
		
		// Make match (40% weight)
		if strings.EqualFold(item.Make, query.Make) {
			score += 0.4
		}
		
		// Model match (40% weight)
		if strings.EqualFold(item.Model, query.Model) {
			score += 0.4
		}
		
		// Year match (20% weight)
		if query.Year != nil && item.Year == *query.Year {
			score += 0.2
		} else if query.Year == nil {
			score += 0.1 // Partial credit if no year specified
		}
		
		totalScore += score
	}
	
	return totalScore / float64(len(items))
}