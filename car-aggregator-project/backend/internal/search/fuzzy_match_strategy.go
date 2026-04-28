package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"car-aggregator/internal/config"
	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// FuzzyMatchStrategy implements fuzzy matching for partial and approximate searches
type FuzzyMatchStrategy struct {
	*BaseStrategy
	carAPIClient   CarAPIClient
	config         *config.SearchConfiguration
	popularModels  []string
	makeVariations map[string][]string
	modelVariations map[string][]string
}

// NewFuzzyMatchStrategy creates a new fuzzy match strategy
func NewFuzzyMatchStrategy(
	carAPIClient CarAPIClient,
	config *config.SearchConfiguration,
	logger logging.SearchLogger,
) *FuzzyMatchStrategy {
	return &FuzzyMatchStrategy{
		BaseStrategy:    NewBaseStrategy(StrategyNameFuzzyMatch, PriorityFuzzyMatch, logger),
		carAPIClient:    carAPIClient,
		config:          config,
		popularModels:   config.Fallback.PopularModels,
		makeVariations:  initMakeVariations(),
		modelVariations: initModelVariations(),
	}
}

// CanHandle determines if this strategy can handle the given query
func (fms *FuzzyMatchStrategy) CanHandle(query *ParsedQuery) bool {
	// Can handle queries with low to medium confidence
	// or when exact strategies haven't found good results
	return query.Confidence < 0.8
}

// Search performs fuzzy matching search
func (fms *FuzzyMatchStrategy) Search(ctx context.Context, query *ParsedQuery) (*SearchResult, error) {
	fms.LogStart(query)
	startTime := time.Now()
	
	var allItems []dtos.SearchTrimItem
	var bestConfidence float64
	
	// Try different fuzzy matching approaches
	approaches := []func(context.Context, *ParsedQuery) ([]dtos.SearchTrimItem, float64, error){
		fms.fuzzyMakeModelSearch,
		fms.fuzzyModelOnlySearch,
		fms.popularModelSearch,
		fms.keywordSearch,
	}
	
	for _, approach := range approaches {
		items, confidence, err := approach(ctx, query)
		if err != nil {
			fms.logger.LogError("fuzzy_approach_failed", err, map[string]interface{}{
				"query": query.Original,
			})
			continue
		}
		
		if len(items) > 0 {
			allItems = append(allItems, items...)
			if confidence > bestConfidence {
				bestConfidence = confidence
			}
		}
	}
	
	duration := time.Since(startTime)
	
	if len(allItems) == 0 {
		return &SearchResult{
			Items:      []dtos.SearchTrimItem{},
			Source:     fms.Name(),
			Confidence: 0.0,
			Duration:   duration,
		}, nil
	}
	
	// Remove duplicates and rank results
	uniqueItems := fms.removeDuplicates(allItems)
	rankedItems := fms.rankResults(query, uniqueItems)
	
	// Limit results
	maxResults := 15
	if len(rankedItems) > maxResults {
		rankedItems = rankedItems[:maxResults]
	}
	
	finalResult := &SearchResult{
		Items:      rankedItems,
		Source:     fms.Name(),
		Confidence: bestConfidence,
		Duration:   duration,
		Metadata: map[string]interface{}{
			"total_found":    len(allItems),
			"unique_items":   len(uniqueItems),
			"final_results":  len(rankedItems),
			"search_type":    "fuzzy_match",
		},
	}
	
	fms.LogResult(query, finalResult, nil)
	return finalResult, nil
}

// fuzzyMakeModelSearch performs fuzzy matching on make and model
func (fms *FuzzyMatchStrategy) fuzzyMakeModelSearch(ctx context.Context, query *ParsedQuery) ([]dtos.SearchTrimItem, float64, error) {
	if query.Make == "" || query.Model == "" {
		return nil, 0.0, nil
	}
	
	// Find similar makes and models
	similarMakes := fms.findSimilarMakes(query.Make)
	similarModels := fms.findSimilarModels(query.Model)
	
	var allItems []dtos.SearchTrimItem
	
	// Try combinations of similar makes and models
	for _, make := range similarMakes {
		for _, model := range similarModels {
			response, err := fms.carAPIClient.GetTrimsByMakeAndModel(ctx, make, model, 5)
			if err != nil || response == nil || len(response.Data) == 0 {
				continue
			}
			
			items := fms.convertToSearchItems(response.Data, query)
			allItems = append(allItems, items...)
		}
	}
	
	if len(allItems) == 0 {
		return nil, 0.0, nil
	}
	
	// Calculate confidence based on similarity
	confidence := fms.calculateFuzzyConfidence(query.Make, query.Model, similarMakes[0], similarModels[0])
	
	return allItems, confidence, nil
}

// fuzzyModelOnlySearch performs fuzzy matching on model only
func (fms *FuzzyMatchStrategy) fuzzyModelOnlySearch(ctx context.Context, query *ParsedQuery) ([]dtos.SearchTrimItem, float64, error) {
	if query.Model == "" {
		return nil, 0.0, nil
	}
	
	similarModels := fms.findSimilarModels(query.Model)
	
	var allItems []dtos.SearchTrimItem
	
	for _, model := range similarModels {
		response, err := fms.carAPIClient.GetTrimsByModel(ctx, model, 8)
		if err != nil || response == nil || len(response.Data) == 0 {
			continue
		}
		
		items := fms.convertToSearchItems(response.Data, query)
		allItems = append(allItems, items...)
	}
	
	if len(allItems) == 0 {
		return nil, 0.0, nil
	}
	
	confidence := fms.calculateModelOnlyConfidence(query.Model, similarModels[0])
	
	return allItems, confidence, nil
}

// popularModelSearch searches using popular models
func (fms *FuzzyMatchStrategy) popularModelSearch(ctx context.Context, query *ParsedQuery) ([]dtos.SearchTrimItem, float64, error) {
	queryLower := strings.ToLower(query.Original)
	
	var matchingModels []string
	
	// Find popular models that match the query
	for _, model := range fms.popularModels {
		modelLower := strings.ToLower(model)
		
		// Check if query contains the model name or vice versa
		if strings.Contains(queryLower, modelLower) || strings.Contains(modelLower, queryLower) {
			matchingModels = append(matchingModels, model)
		}
		
		// Check fuzzy similarity
		if fms.calculateStringSimilarity(queryLower, modelLower) > fms.config.Fallback.FuzzyThreshold {
			matchingModels = append(matchingModels, model)
		}
	}
	
	if len(matchingModels) == 0 {
		return nil, 0.0, nil
	}
	
	var allItems []dtos.SearchTrimItem
	
	for _, model := range matchingModels {
		response, err := fms.carAPIClient.GetTrimsByModel(ctx, model, 5)
		if err != nil || response == nil || len(response.Data) == 0 {
			continue
		}
		
		items := fms.convertToSearchItems(response.Data, query)
		allItems = append(allItems, items...)
	}
	
	confidence := 0.6 // Medium confidence for popular model matches
	
	return allItems, confidence, nil
}

// keywordSearch performs keyword-based search
func (fms *FuzzyMatchStrategy) keywordSearch(ctx context.Context, query *ParsedQuery) ([]dtos.SearchTrimItem, float64, error) {
	// Extract keywords from the query
	keywords := fms.extractKeywords(query.Original)
	
	if len(keywords) == 0 {
		return nil, 0.0, nil
	}
	
	var allItems []dtos.SearchTrimItem
	
	// Try each keyword as a potential model
	for _, keyword := range keywords {
		if len(keyword) < 3 { // Skip very short keywords
			continue
		}
		
		response, err := fms.carAPIClient.GetTrimsByModel(ctx, keyword, 3)
		if err != nil || response == nil || len(response.Data) == 0 {
			continue
		}
		
		items := fms.convertToSearchItems(response.Data, query)
		allItems = append(allItems, items...)
	}
	
	confidence := 0.4 // Lower confidence for keyword search
	
	return allItems, confidence, nil
}

// findSimilarMakes finds makes similar to the given make
func (fms *FuzzyMatchStrategy) findSimilarMakes(make string) []string {
	makeLower := strings.ToLower(make)
	
	// Check variations first
	if variations, exists := fms.makeVariations[makeLower]; exists {
		return variations
	}
	
	// Find similar makes using string similarity
	var similar []string
	similar = append(similar, make) // Include original
	
	allMakes := []string{
		"Toyota", "Honda", "Volkswagen", "BMW", "Mercedes-Benz", "Audi",
		"Ford", "Nissan", "Hyundai", "Kia", "Mazda", "Subaru",
		"Chevrolet", "GMC", "Cadillac", "Lexus", "Infiniti", "Acura",
	}
	
	for _, candidate := range allMakes {
		if strings.EqualFold(candidate, make) {
			continue // Skip exact match (already included)
		}
		
		similarity := fms.calculateStringSimilarity(makeLower, strings.ToLower(candidate))
		if similarity > fms.config.Fallback.FuzzyThreshold {
			similar = append(similar, candidate)
		}
	}
	
	return similar
}

// findSimilarModels finds models similar to the given model
func (fms *FuzzyMatchStrategy) findSimilarModels(model string) []string {
	modelLower := strings.ToLower(model)
	
	// Check variations first
	if variations, exists := fms.modelVariations[modelLower]; exists {
		return variations
	}
	
	// Find similar models
	var similar []string
	similar = append(similar, model) // Include original
	
	// Check against popular models
	for _, candidate := range fms.popularModels {
		if strings.EqualFold(candidate, model) {
			continue
		}
		
		similarity := fms.calculateStringSimilarity(modelLower, strings.ToLower(candidate))
		if similarity > fms.config.Fallback.FuzzyThreshold {
			similar = append(similar, candidate)
		}
	}
	
	return similar
}

// calculateStringSimilarity calculates similarity between two strings
func (fms *FuzzyMatchStrategy) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	// Use Levenshtein distance
	distance := fms.levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}
	
	if maxLen == 0 {
		return 1.0
	}
	
	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance calculates Levenshtein distance between two strings
func (fms *FuzzyMatchStrategy) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len(s1)][len(s2)]
}

// calculateFuzzyConfidence calculates confidence for fuzzy make+model match
func (fms *FuzzyMatchStrategy) calculateFuzzyConfidence(originalMake, originalModel, matchedMake, matchedModel string) float64 {
	makeSimilarity := fms.calculateStringSimilarity(strings.ToLower(originalMake), strings.ToLower(matchedMake))
	modelSimilarity := fms.calculateStringSimilarity(strings.ToLower(originalModel), strings.ToLower(matchedModel))
	
	// Weighted average (model is more important)
	confidence := (makeSimilarity*0.4 + modelSimilarity*0.6) * 0.7 // Max 0.7 for fuzzy match
	
	return confidence
}

// calculateModelOnlyConfidence calculates confidence for model-only fuzzy match
func (fms *FuzzyMatchStrategy) calculateModelOnlyConfidence(originalModel, matchedModel string) float64 {
	similarity := fms.calculateStringSimilarity(strings.ToLower(originalModel), strings.ToLower(matchedModel))
	return similarity * 0.6 // Max 0.6 for model-only fuzzy match
}

// extractKeywords extracts meaningful keywords from the query
func (fms *FuzzyMatchStrategy) extractKeywords(query string) []string {
	// Split by spaces and clean
	words := strings.Fields(strings.ToLower(query))
	
	// Filter out common words and short words
	stopWords := map[string]bool{
		"car": true, "auto": true, "vehicle": true, "the": true, "a": true, "an": true,
		"and": true, "or": true, "but": true, "in": true, "on": true, "at": true,
		"to": true, "for": true, "of": true, "with": true, "by": true,
	}
	
	var keywords []string
	for _, word := range words {
		if len(word) >= 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// convertToSearchItems converts CarAPI trims to SearchTrimItems
func (fms *FuzzyMatchStrategy) convertToSearchItems(trims []dtos.Trim, query *ParsedQuery) []dtos.SearchTrimItem {
	items := make([]dtos.SearchTrimItem, 0, len(trims))
	
	for _, trim := range trims {
		// Get additional specs
		seats, fuel, _ := fms.carAPIClient.GetVehicleSpecs(context.Background(), trim.Year, trim.Make, trim.Model)
		
		// Infer transmission
		transmission := fms.inferTransmission(trim.Description, trim.Trim)
		
		// Generate image URL
		imageURL := fms.buildImageURL(trim.Year, trim.Make, trim.Model, trim.Trim)
		
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
		
		// Apply year filter if specified
		if query.Year != nil && trim.Year != *query.Year {
			continue
		}
		
		items = append(items, item)
	}
	
	return items
}

// rankResults ranks search results by relevance to the query
func (fms *FuzzyMatchStrategy) rankResults(query *ParsedQuery, items []dtos.SearchTrimItem) []dtos.SearchTrimItem {
	type scoredItem struct {
		item  dtos.SearchTrimItem
		score float64
	}
	
	var scored []scoredItem
	
	for _, item := range items {
		score := fms.calculateRelevanceScore(query, item)
		scored = append(scored, scoredItem{item: item, score: score})
	}
	
	// Sort by score (highest first)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	
	// Extract items
	var ranked []dtos.SearchTrimItem
	for _, s := range scored {
		ranked = append(ranked, s.item)
	}
	
	return ranked
}

// calculateRelevanceScore calculates relevance score for an item
func (fms *FuzzyMatchStrategy) calculateRelevanceScore(query *ParsedQuery, item dtos.SearchTrimItem) float64 {
	score := 0.0
	
	// Make similarity (30% weight)
	if query.Make != "" {
		makeSimilarity := fms.calculateStringSimilarity(strings.ToLower(query.Make), strings.ToLower(item.Make))
		score += makeSimilarity * 0.3
	}
	
	// Model similarity (40% weight)
	if query.Model != "" {
		modelSimilarity := fms.calculateStringSimilarity(strings.ToLower(query.Model), strings.ToLower(item.Model))
		score += modelSimilarity * 0.4
	}
	
	// Year match (20% weight)
	if query.Year != nil {
		if item.Year == *query.Year {
			score += 0.2
		} else {
			// Partial credit for nearby years
			yearDiff := abs(item.Year - *query.Year)
			if yearDiff <= 2 {
				score += 0.1
			}
		}
	} else {
		score += 0.1 // Partial credit if no year specified
	}
	
	// Popularity bonus (10% weight) - newer cars and popular models
	if item.Year >= 2015 {
		score += 0.05
	}
	
	// Popular model bonus
	for _, popular := range fms.popularModels {
		if strings.EqualFold(item.Model, popular) {
			score += 0.05
			break
		}
	}
	
	return score
}

// removeDuplicates removes duplicate items
func (fms *FuzzyMatchStrategy) removeDuplicates(items []dtos.SearchTrimItem) []dtos.SearchTrimItem {
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

// Helper functions
func (fms *FuzzyMatchStrategy) inferTransmission(description string, trimName string) string {
	combined := strings.ToLower(description + " " + trimName)
	
	if strings.Contains(combined, "cvt") || strings.Contains(combined, "automatic") || strings.Contains(combined, " auto") {
		return "Automatic"
	}
	
	if strings.Contains(combined, "manual") || strings.Contains(combined, " 5m") || strings.Contains(combined, " 6m") {
		return "Manual"
	}
	
	return "Manual" // Default
}

func (fms *FuzzyMatchStrategy) buildImageURL(year int, make, model, trim string) string {
	query := fmt.Sprintf("%d %s %s", year, make, model)
	encoded := strings.ReplaceAll(strings.ToLower(query), " ", "_")
	return fmt.Sprintf("https://www.regcheck.org.uk/image.aspx/@%s", encoded)
}

func initMakeVariations() map[string][]string {
	return map[string][]string{
		"vw":         {"Volkswagen", "VW"},
		"volkswagen": {"Volkswagen", "VW"},
		"bmw":        {"BMW"},
		"mercedes":   {"Mercedes-Benz", "Mercedes", "Benz"},
		"benz":       {"Mercedes-Benz", "Mercedes", "Benz"},
		"mb":         {"Mercedes-Benz", "Mercedes"},
		"chevy":      {"Chevrolet", "Chevy"},
		"chevrolet":  {"Chevrolet", "Chevy"},
	}
}

func initModelVariations() map[string][]string {
	return map[string][]string{
		"golf":    {"Golf"},
		"camry":   {"Camry"},
		"civic":   {"Civic"},
		"corolla": {"Corolla"},
		"focus":   {"Focus"},
		"accord":  {"Accord"},
		"cr-v":    {"CR-V", "CRV"},
		"crv":     {"CR-V", "CRV"},
		"f-150":   {"F-150", "F150"},
		"f150":    {"F-150", "F150"},
		"rav4":    {"RAV4", "Rav4"},
	}
}