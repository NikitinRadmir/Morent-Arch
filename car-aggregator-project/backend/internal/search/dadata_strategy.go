package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// DaDataStrategy implements search using DaData API for query normalization
type DaDataStrategy struct {
	*BaseStrategy
	dadataClient DaDataClient
	carAPIClient CarAPIClient
}

// NewDaDataStrategy creates a new DaData strategy
func NewDaDataStrategy(
	dadataClient DaDataClient,
	carAPIClient CarAPIClient,
	logger logging.SearchLogger,
) *DaDataStrategy {
	return &DaDataStrategy{
		BaseStrategy: NewBaseStrategy(StrategyNameDaData, PriorityDaData, logger),
		dadataClient: dadataClient,
		carAPIClient: carAPIClient,
	}
}

// CanHandle determines if this strategy can handle the given query
func (ds *DaDataStrategy) CanHandle(query *ParsedQuery) bool {
	// Can handle most queries, but prefer exact match for high-confidence queries
	return query.Confidence < 0.9
}

// Search performs search using DaData for normalization and CarAPI for results
func (ds *DaDataStrategy) Search(ctx context.Context, query *ParsedQuery) (*SearchResult, error) {
	ds.LogStart(query)
	startTime := time.Now()
	
	// Use DaData to normalize the query
	dadataResults, err := ds.dadataClient.SearchVehicle(ctx, query.Original)
	if err != nil {
		ds.LogResult(query, nil, err)
		return nil, fmt.Errorf("DaData search failed: %w", err)
	}
	
	if len(dadataResults) == 0 {
		return &SearchResult{
			Items:      []dtos.SearchTrimItem{},
			Source:     ds.Name(),
			Confidence: 0.0,
		}, nil
	}
	
	// Try multiple DaData results to find the best match
	var bestResult *SearchResult
	var allItems []dtos.SearchTrimItem
	
	for i, dadataResult := range dadataResults {
		// Limit to top 3 DaData results to avoid too many API calls
		if i >= 3 {
			break
		}
		
		result, err := ds.searchWithDaDataResult(ctx, query, dadataResult)
		if err != nil {
			ds.logger.LogError("dadata_result_search_failed", err, map[string]interface{}{
				"dadata_result": dadataResult,
				"query":         query.Original,
			})
			continue
		}
		
		if result != nil && len(result.Items) > 0 {
			allItems = append(allItems, result.Items...)
			
			if bestResult == nil || result.Confidence > bestResult.Confidence {
				bestResult = result
			}
		}
	}
	
	duration := time.Since(startTime)
	
	if bestResult == nil {
		return &SearchResult{
			Items:      []dtos.SearchTrimItem{},
			Source:     ds.Name(),
			Confidence: 0.0,
			Duration:   duration,
		}, nil
	}
	
	// Combine results and remove duplicates
	uniqueItems := ds.removeDuplicates(allItems)
	
	finalResult := &SearchResult{
		Items:      uniqueItems,
		Source:     ds.Name(),
		Confidence: ds.calculateOverallConfidence(dadataResults, bestResult.Confidence),
		Duration:   duration,
		Metadata: map[string]interface{}{
			"dadata_results_count": len(dadataResults),
			"best_dadata_result":   dadataResults[0],
			"total_items":          len(uniqueItems),
		},
	}
	
	ds.LogResult(query, finalResult, nil)
	return finalResult, nil
}

// searchWithDaDataResult performs search using a specific DaData result
func (ds *DaDataStrategy) searchWithDaDataResult(ctx context.Context, query *ParsedQuery, dadataResult VehicleResult) (*SearchResult, error) {
	make := dadataResult.Brand
	model := dadataResult.Model
	
	if make == "" || model == "" {
		return nil, fmt.Errorf("incomplete DaData result: make=%s, model=%s", make, model)
	}
	
	// Try make+model search first
	response, err := ds.carAPIClient.GetTrimsByMakeAndModel(ctx, make, model, 10)
	if err != nil || response == nil || len(response.Data) == 0 {
		// Fallback to model-only search
		response, err = ds.carAPIClient.GetTrimsByModel(ctx, model, 10)
		if err != nil {
			return nil, fmt.Errorf("CarAPI search failed for make=%s, model=%s: %w", make, model, err)
		}
	}
	
	if response == nil || len(response.Data) == 0 {
		return &SearchResult{
			Items:      []dtos.SearchTrimItem{},
			Source:     ds.Name(),
			Confidence: 0.0,
		}, nil
	}
	
	// Convert to search items
	items := ds.convertToSearchItems(response.Data, query, dadataResult)
	
	// Calculate confidence based on DaData confidence and result quality
	confidence := ds.calculateResultConfidence(dadataResult, items, query)
	
	return &SearchResult{
		Items:      items,
		Source:     ds.Name(),
		Confidence: confidence,
		Metadata: map[string]interface{}{
			"dadata_make":       make,
			"dadata_model":      model,
			"dadata_confidence": dadataResult.Confidence,
			"carapi_results":    len(response.Data),
		},
	}, nil
}

// convertToSearchItems converts CarAPI trims to SearchTrimItems with DaData context
func (ds *DaDataStrategy) convertToSearchItems(trims []dtos.Trim, query *ParsedQuery, dadataResult VehicleResult) []dtos.SearchTrimItem {
	items := make([]dtos.SearchTrimItem, 0, len(trims))
	
	for _, trim := range trims {
		// Get additional specs
		seats, fuel, _ := ds.carAPIClient.GetVehicleSpecs(context.Background(), trim.Year, trim.Make, trim.Model)
		
		// Infer transmission
		transmission := ds.inferTransmission(trim.Description, trim.Trim)
		
		// Generate image URL
		imageURL := ds.buildImageURL(trim.Year, trim.Make, trim.Model, trim.Trim)
		
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
		
		// Apply filters
		if ds.shouldIncludeItem(item, query, dadataResult) {
			items = append(items, item)
		}
	}
	
	return items
}

// shouldIncludeItem determines if an item should be included in results
func (ds *DaDataStrategy) shouldIncludeItem(item dtos.SearchTrimItem, query *ParsedQuery, dadataResult VehicleResult) bool {
	// Year filter
	if query.Year != nil && item.Year != *query.Year {
		return false
	}
	
	// Make filter - should match DaData result or original query
	if query.Make != "" {
		makeMatch := strings.EqualFold(item.Make, query.Make) ||
			strings.EqualFold(item.Make, dadataResult.Brand)
		if !makeMatch {
			return false
		}
	}
	
	// Model filter - should match DaData result or original query
	if query.Model != "" {
		modelMatch := strings.EqualFold(item.Model, query.Model) ||
			strings.EqualFold(item.Model, dadataResult.Model)
		if !modelMatch {
			return false
		}
	}
	
	return true
}

// calculateResultConfidence calculates confidence for a specific result
func (ds *DaDataStrategy) calculateResultConfidence(dadataResult VehicleResult, items []dtos.SearchTrimItem, query *ParsedQuery) float64 {
	baseConfidence := dadataResult.Confidence * 0.8 // Start with DaData confidence, slightly reduced
	
	if len(items) == 0 {
		return 0.0
	}
	
	// Boost confidence based on result quality
	if len(items) >= 5 {
		baseConfidence += 0.1
	}
	
	// Check if results match the original query well
	if len(items) > 0 {
		firstItem := items[0]
		
		// Make match bonus
		if query.Make != "" && strings.EqualFold(firstItem.Make, query.Make) {
			baseConfidence += 0.05
		}
		
		// Model match bonus
		if query.Model != "" && strings.EqualFold(firstItem.Model, query.Model) {
			baseConfidence += 0.05
		}
		
		// Year match bonus
		if query.Year != nil && firstItem.Year == *query.Year {
			baseConfidence += 0.05
		}
	}
	
	// Cap at 0.95 (leave room for exact match strategy)
	if baseConfidence > 0.95 {
		baseConfidence = 0.95
	}
	
	return baseConfidence
}

// calculateOverallConfidence calculates overall confidence from multiple DaData results
func (ds *DaDataStrategy) calculateOverallConfidence(dadataResults []VehicleResult, bestResultConfidence float64) float64 {
	if len(dadataResults) == 0 {
		return 0.0
	}
	
	// Use the best result confidence as base
	confidence := bestResultConfidence
	
	// Boost if we have multiple good DaData results
	goodResults := 0
	for _, result := range dadataResults {
		if result.Confidence > 0.6 {
			goodResults++
		}
	}
	
	if goodResults > 1 {
		confidence += 0.05
	}
	
	// Cap at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}
	
	return confidence
}

// removeDuplicates removes duplicate items based on make, model, year, and trim
func (ds *DaDataStrategy) removeDuplicates(items []dtos.SearchTrimItem) []dtos.SearchTrimItem {
	seen := make(map[string]bool)
	var unique []dtos.SearchTrimItem
	
	for _, item := range items {
		key := fmt.Sprintf("%s_%s_%d_%s", 
			strings.ToLower(item.Make), 
			strings.ToLower(item.Model), 
			item.Year, 
			strings.ToLower(item.Trim))
		
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}
	
	return unique
}

// inferTransmission infers transmission type from description and trim
func (ds *DaDataStrategy) inferTransmission(description string, trimName string) string {
	combined := strings.ToLower(description + " " + trimName)
	
	// Automatic indicators
	automaticIndicators := []string{
		"cvt", "automatic", " auto", " 6a", " 7a", " 8a", " 9a", " 10a",
		"6-speed automatic", "7-speed automatic", "8-speed automatic",
		"continuously variable", "dual-clutch", "dct", "tiptronic",
	}
	
	for _, indicator := range automaticIndicators {
		if strings.Contains(combined, indicator) {
			return "Automatic"
		}
	}
	
	// Manual indicators
	manualIndicators := []string{
		"manual", " 5m", " 6m", "5-speed manual", "6-speed manual",
		"stick", "mt", "standard transmission",
	}
	
	for _, indicator := range manualIndicators {
		if strings.Contains(combined, indicator) {
			return "Manual"
		}
	}
	
	// Default based on common patterns
	if strings.Contains(combined, "sedan") || strings.Contains(combined, "suv") {
		return "Automatic" // Most sedans and SUVs are automatic
	}
	
	return "Manual" // Conservative default
}

// buildImageURL builds an image URL for the vehicle
func (ds *DaDataStrategy) buildImageURL(year int, make, model, trim string) string {
	query := strings.TrimSpace(fmt.Sprintf("%d %s %s", year, make, model))
	if query == "" {
		return ""
	}
	
	// Simple encoding for image service
	encoded := strings.ReplaceAll(strings.ToLower(query), " ", "_")
	return fmt.Sprintf("https://www.regcheck.org.uk/image.aspx/@%s", encoded)
}

// GetDaDataResults returns the DaData results for debugging
func (ds *DaDataStrategy) GetDaDataResults(ctx context.Context, query string) ([]VehicleResult, error) {
	return ds.dadataClient.SearchVehicle(ctx, query)
}

// ValidateResults validates that the results match the DaData normalization
func (ds *DaDataStrategy) ValidateResults(dadataResults []VehicleResult, items []dtos.SearchTrimItem) bool {
	if len(dadataResults) == 0 || len(items) == 0 {
		return false
	}
	
	// Check if at least one item matches any DaData result
	for _, dadataResult := range dadataResults {
		for _, item := range items {
			makeMatch := strings.EqualFold(item.Make, dadataResult.Brand)
			modelMatch := strings.EqualFold(item.Model, dadataResult.Model)
			
			if makeMatch && modelMatch {
				return true
			}
		}
	}
	
	return false
}

// GetSearchQuality returns a quality assessment of the search results
func (ds *DaDataStrategy) GetSearchQuality(query *ParsedQuery, dadataResults []VehicleResult, items []dtos.SearchTrimItem) map[string]interface{} {
	quality := map[string]interface{}{
		"dadata_results_count": len(dadataResults),
		"items_count":          len(items),
		"has_exact_match":      false,
		"has_partial_match":    false,
		"avg_dadata_confidence": 0.0,
	}
	
	if len(dadataResults) > 0 {
		totalConfidence := 0.0
		for _, result := range dadataResults {
			totalConfidence += result.Confidence
		}
		quality["avg_dadata_confidence"] = totalConfidence / float64(len(dadataResults))
	}
	
	// Check for matches
	for _, item := range items {
		if query.Make != "" && query.Model != "" {
			makeMatch := strings.EqualFold(item.Make, query.Make)
			modelMatch := strings.EqualFold(item.Model, query.Model)
			
			if makeMatch && modelMatch {
				quality["has_exact_match"] = true
			} else if makeMatch || modelMatch {
				quality["has_partial_match"] = true
			}
		}
	}
	
	return quality
}