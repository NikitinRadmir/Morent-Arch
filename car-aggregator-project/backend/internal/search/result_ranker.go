package search

import (
	"math"
	"sort"
	"strings"
	"time"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// ResultRanker interface for ranking search results
type ResultRanker interface {
	RankResults(query *ParsedQuery, results []dtos.SearchTrimItem) []dtos.SearchTrimItem
	CalculateRelevanceScore(query *ParsedQuery, item dtos.SearchTrimItem) float64
	SetCriteria(criteria RankingCriteria)
	GetCriteria() RankingCriteria
}

// RankingCriteria defines weights for different ranking factors
type RankingCriteria struct {
	MakeMatch     float64 `json:"make_match"`     // Weight for make matching
	ModelMatch    float64 `json:"model_match"`    // Weight for model matching
	YearProximity float64 `json:"year_proximity"` // Weight for year proximity
	Popularity    float64 `json:"popularity"`     // Weight for model popularity
	Price         float64 `json:"price"`          // Weight for price considerations
	Recency       float64 `json:"recency"`        // Weight for newer models
	Completeness  float64 `json:"completeness"`   // Weight for data completeness
}

// DefaultRankingCriteria returns default ranking criteria
func DefaultRankingCriteria() RankingCriteria {
	return RankingCriteria{
		MakeMatch:     0.25, // 25%
		ModelMatch:    0.30, // 30%
		YearProximity: 0.15, // 15%
		Popularity:    0.10, // 10%
		Price:         0.05, // 5%
		Recency:       0.10, // 10%
		Completeness:  0.05, // 5%
	}
}

// resultRanker implements the ResultRanker interface
type resultRanker struct {
	criteria      RankingCriteria
	logger        logging.SearchLogger
	popularModels map[string]int // Model name -> popularity score
	popularMakes  map[string]int // Make name -> popularity score
}

// NewResultRanker creates a new result ranker
func NewResultRanker(logger logging.SearchLogger) ResultRanker {
	return &resultRanker{
		criteria:      DefaultRankingCriteria(),
		logger:        logger,
		popularModels: initPopularModels(),
		popularMakes:  initPopularMakes(),
	}
}

// RankResults ranks the search results based on relevance to the query
func (rr *resultRanker) RankResults(query *ParsedQuery, results []dtos.SearchTrimItem) []dtos.SearchTrimItem {
	if len(results) == 0 {
		return results
	}
	
	// Create scored items
	type scoredItem struct {
		item  dtos.SearchTrimItem
		score float64
	}
	
	var scoredItems []scoredItem
	
	for _, item := range results {
		score := rr.CalculateRelevanceScore(query, item)
		scoredItems = append(scoredItems, scoredItem{
			item:  item,
			score: score,
		})
	}
	
	// Sort by score (highest first)
	sort.Slice(scoredItems, func(i, j int) bool {
		return scoredItems[i].score > scoredItems[j].score
	})
	
	// Extract ranked items
	rankedResults := make([]dtos.SearchTrimItem, len(scoredItems))
	for i, scored := range scoredItems {
		rankedResults[i] = scored.item
	}
	
	rr.logger.LogSearchComplete(query.Original, len(rankedResults), 0)
	
	return rankedResults
}

// CalculateRelevanceScore calculates the relevance score for a single item
func (rr *resultRanker) CalculateRelevanceScore(query *ParsedQuery, item dtos.SearchTrimItem) float64 {
	var totalScore float64
	
	// Make match score
	makeScore := rr.calculateMakeMatchScore(query, item)
	totalScore += makeScore * rr.criteria.MakeMatch
	
	// Model match score
	modelScore := rr.calculateModelMatchScore(query, item)
	totalScore += modelScore * rr.criteria.ModelMatch
	
	// Year proximity score
	yearScore := rr.calculateYearProximityScore(query, item)
	totalScore += yearScore * rr.criteria.YearProximity
	
	// Popularity score
	popularityScore := rr.calculatePopularityScore(item)
	totalScore += popularityScore * rr.criteria.Popularity
	
	// Price score
	priceScore := rr.calculatePriceScore(item)
	totalScore += priceScore * rr.criteria.Price
	
	// Recency score
	recencyScore := rr.calculateRecencyScore(item)
	totalScore += recencyScore * rr.criteria.Recency
	
	// Completeness score
	completenessScore := rr.calculateCompletenessScore(item)
	totalScore += completenessScore * rr.criteria.Completeness
	
	return totalScore
}

// calculateMakeMatchScore calculates score based on make matching
func (rr *resultRanker) calculateMakeMatchScore(query *ParsedQuery, item dtos.SearchTrimItem) float64 {
	if query.Make == "" {
		return 0.5 // Neutral score if no make specified
	}
	
	queryMake := strings.ToLower(strings.TrimSpace(query.Make))
	itemMake := strings.ToLower(strings.TrimSpace(item.Make))
	
	// Exact match
	if queryMake == itemMake {
		return 1.0
	}
	
	// Check for common abbreviations and variations
	if rr.areMakeVariations(queryMake, itemMake) {
		return 0.9
	}
	
	// Fuzzy string matching
	similarity := rr.calculateStringSimilarity(queryMake, itemMake)
	if similarity > 0.8 {
		return similarity
	}
	
	return 0.0
}

// calculateModelMatchScore calculates score based on model matching
func (rr *resultRanker) calculateModelMatchScore(query *ParsedQuery, item dtos.SearchTrimItem) float64 {
	if query.Model == "" {
		return 0.5 // Neutral score if no model specified
	}
	
	queryModel := strings.ToLower(strings.TrimSpace(query.Model))
	itemModel := strings.ToLower(strings.TrimSpace(item.Model))
	
	// Exact match
	if queryModel == itemModel {
		return 1.0
	}
	
	// Check for common variations
	if rr.areModelVariations(queryModel, itemModel) {
		return 0.9
	}
	
	// Partial match (one contains the other)
	if strings.Contains(queryModel, itemModel) || strings.Contains(itemModel, queryModel) {
		return 0.8
	}
	
	// Fuzzy string matching
	similarity := rr.calculateStringSimilarity(queryModel, itemModel)
	if similarity > 0.7 {
		return similarity
	}
	
	return 0.0
}

// calculateYearProximityScore calculates score based on year proximity
func (rr *resultRanker) calculateYearProximityScore(query *ParsedQuery, item dtos.SearchTrimItem) float64 {
	if query.Year == nil {
		// If no year specified, prefer newer models
		currentYear := time.Now().Year()
		yearDiff := currentYear - item.Year
		
		if yearDiff <= 2 {
			return 1.0 // Very recent
		} else if yearDiff <= 5 {
			return 0.8 // Recent
		} else if yearDiff <= 10 {
			return 0.6 // Moderately old
		} else {
			return 0.3 // Old
		}
	}
	
	// Exact year match
	if item.Year == *query.Year {
		return 1.0
	}
	
	// Calculate proximity score
	yearDiff := abs(item.Year - *query.Year)
	
	if yearDiff == 1 {
		return 0.8
	} else if yearDiff == 2 {
		return 0.6
	} else if yearDiff <= 5 {
		return 0.4
	} else {
		return 0.1
	}
}

// calculatePopularityScore calculates score based on make/model popularity
func (rr *resultRanker) calculatePopularityScore(item dtos.SearchTrimItem) float64 {
	makeScore := 0.0
	modelScore := 0.0
	
	// Make popularity
	if score, exists := rr.popularMakes[strings.ToLower(item.Make)]; exists {
		makeScore = float64(score) / 100.0 // Normalize to 0-1
	}
	
	// Model popularity
	if score, exists := rr.popularModels[strings.ToLower(item.Model)]; exists {
		modelScore = float64(score) / 100.0 // Normalize to 0-1
	}
	
	// Weighted average (model is more important for popularity)
	return (makeScore*0.3 + modelScore*0.7)
}

// calculatePriceScore calculates score based on price reasonableness
func (rr *resultRanker) calculatePriceScore(item dtos.SearchTrimItem) float64 {
	if item.MSRP <= 0 {
		return 0.5 // Neutral if no price info
	}
	
	// Score based on price ranges (prefer mid-range prices)
	if item.MSRP >= 20000 && item.MSRP <= 50000 {
		return 1.0 // Sweet spot
	} else if item.MSRP >= 15000 && item.MSRP <= 70000 {
		return 0.8 // Reasonable range
	} else if item.MSRP >= 10000 && item.MSRP <= 100000 {
		return 0.6 // Acceptable range
	} else {
		return 0.3 // Too cheap or too expensive
	}
}

// calculateRecencyScore calculates score based on model year recency
func (rr *resultRanker) calculateRecencyScore(item dtos.SearchTrimItem) float64 {
	currentYear := time.Now().Year()
	yearDiff := currentYear - item.Year
	
	if yearDiff <= 1 {
		return 1.0 // Current or last year
	} else if yearDiff <= 3 {
		return 0.8 // Very recent
	} else if yearDiff <= 5 {
		return 0.6 // Recent
	} else if yearDiff <= 10 {
		return 0.4 // Moderately old
	} else {
		return 0.2 // Old
	}
}

// calculateCompletenessScore calculates score based on data completeness
func (rr *resultRanker) calculateCompletenessScore(item dtos.SearchTrimItem) float64 {
	score := 0.0
	totalFields := 8.0 // Total number of fields to check
	
	// Check each field for completeness
	if item.Make != "" {
		score += 1.0
	}
	if item.Model != "" {
		score += 1.0
	}
	if item.Year > 0 {
		score += 1.0
	}
	if item.Trim != "" {
		score += 1.0
	}
	if item.Description != "" {
		score += 1.0
	}
	if item.MSRP > 0 {
		score += 1.0
	}
	if item.Transmission != "" {
		score += 1.0
	}
	if item.ImageURL != "" {
		score += 1.0
	}
	
	return score / totalFields
}

// calculateStringSimilarity calculates similarity between two strings using Levenshtein distance
func (rr *resultRanker) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	distance := rr.levenshteinDistance(s1, s2)
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
func (rr *resultRanker) levenshteinDistance(s1, s2 string) int {
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

// areMakeVariations checks if two makes are variations of each other
func (rr *resultRanker) areMakeVariations(make1, make2 string) bool {
	variations := map[string][]string{
		"vw":         {"volkswagen"},
		"volkswagen": {"vw"},
		"bmw":        {"bmw"},
		"mercedes":   {"mercedes-benz", "benz", "mb"},
		"benz":       {"mercedes-benz", "mercedes", "mb"},
		"mb":         {"mercedes-benz", "mercedes", "benz"},
		"chevy":      {"chevrolet"},
		"chevrolet":  {"chevy"},
	}
	
	if vars, exists := variations[make1]; exists {
		for _, variant := range vars {
			if variant == make2 {
				return true
			}
		}
	}
	
	if vars, exists := variations[make2]; exists {
		for _, variant := range vars {
			if variant == make1 {
				return true
			}
		}
	}
	
	return false
}

// areModelVariations checks if two models are variations of each other
func (rr *resultRanker) areModelVariations(model1, model2 string) bool {
	variations := map[string][]string{
		"cr-v":   {"crv"},
		"crv":    {"cr-v"},
		"f-150":  {"f150"},
		"f150":   {"f-150"},
		"rav4":   {"rav-4"},
		"rav-4":  {"rav4"},
	}
	
	if vars, exists := variations[model1]; exists {
		for _, variant := range vars {
			if variant == model2 {
				return true
			}
		}
	}
	
	if vars, exists := variations[model2]; exists {
		for _, variant := range vars {
			if variant == model1 {
				return true
			}
		}
	}
	
	return false
}

// SetCriteria sets the ranking criteria
func (rr *resultRanker) SetCriteria(criteria RankingCriteria) {
	// Validate that weights sum to approximately 1.0
	total := criteria.MakeMatch + criteria.ModelMatch + criteria.YearProximity +
		criteria.Popularity + criteria.Price + criteria.Recency + criteria.Completeness
	
	if math.Abs(total-1.0) > 0.01 {
		// Normalize weights if they don't sum to 1.0
		criteria.MakeMatch /= total
		criteria.ModelMatch /= total
		criteria.YearProximity /= total
		criteria.Popularity /= total
		criteria.Price /= total
		criteria.Recency /= total
		criteria.Completeness /= total
	}
	
	rr.criteria = criteria
}

// GetCriteria returns the current ranking criteria
func (rr *resultRanker) GetCriteria() RankingCriteria {
	return rr.criteria
}

// Initialize popularity maps
func initPopularMakes() map[string]int {
	return map[string]int{
		"toyota":     95,
		"honda":      90,
		"ford":       85,
		"chevrolet":  80,
		"nissan":     75,
		"hyundai":    70,
		"kia":        65,
		"volkswagen": 70,
		"bmw":        75,
		"mercedes":   70,
		"audi":       65,
		"mazda":      60,
		"subaru":     55,
		"lexus":      60,
		"acura":      55,
		"infiniti":   50,
	}
}

func initPopularModels() map[string]int {
	return map[string]int{
		"camry":      95,
		"corolla":    90,
		"civic":      90,
		"accord":     85,
		"altima":     80,
		"sentra":     75,
		"elantra":    70,
		"sonata":     65,
		"golf":       70,
		"jetta":      65,
		"passat":     60,
		"focus":      70,
		"fusion":     65,
		"fiesta":     60,
		"cruze":      65,
		"malibu":     60,
		"impala":     55,
		"rav4":       85,
		"cr-v":       85,
		"rogue":      80,
		"escape":     75,
		"equinox":    70,
		"tucson":     65,
		"sportage":   60,
		"cx-5":       65,
		"outback":    60,
		"forester":   55,
		"x3":         60,
		"x5":         65,
		"q5":         55,
		"glc":        55,
	}
}