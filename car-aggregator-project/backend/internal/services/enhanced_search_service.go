package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"car-aggregator/internal/config"
	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
	"car-aggregator/internal/repositories"
	"car-aggregator/internal/search"
	"car-aggregator/internal/validators"
)

// EnhancedSearchService provides improved search functionality
type EnhancedSearchService interface {
	SearchTrims(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error)
	SearchTrimsWithOptions(ctx context.Context, query dtos.UserSearchQuery, options SearchOptions) (EnhancedSearchTrimsResponse, error)
	SearchWithFallback(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error)
}

// SearchOptions provides additional search configuration
type SearchOptions struct {
	MaxResults      int           `json:"max_results"`
	EnableFallback  bool          `json:"enable_fallback"`
	IncludeDebugInfo bool         `json:"include_debug_info"`
	Timeout         time.Duration `json:"timeout"`
	Strategy        string        `json:"strategy"` // "auto", "exact", "dadata", "fuzzy"
}

// EnhancedSearchTrimsResponse provides detailed search results
type EnhancedSearchTrimsResponse struct {
	Query       string                    `json:"query"`
	Count       int                       `json:"count"`
	Cars        []dtos.SearchTrimItem     `json:"cars"`
	DebugInfo   *SearchDebugInfo          `json:"debug_info,omitempty"`
	Suggestions []string                  `json:"suggestions,omitempty"`
	Metadata    map[string]interface{}    `json:"metadata,omitempty"`
}

// SearchDebugInfo provides debugging information
type SearchDebugInfo struct {
	ParsedQuery     *search.ParsedQuery      `json:"parsed_query"`
	StrategiesUsed  []string                 `json:"strategies_used"`
	ProcessingTime  time.Duration            `json:"processing_time"`
	ValidationInfo  map[string]interface{}   `json:"validation_info,omitempty"`
	RankingInfo     map[string]interface{}   `json:"ranking_info,omitempty"`
}

// enhancedSearchService implements the EnhancedSearchService interface
type enhancedSearchService struct {
	// Core components
	queryProcessor   search.QueryProcessor
	queryValidator   search.QueryValidator
	strategyChain    *search.StrategyChain
	resultRanker     search.ResultRanker
	logger           logging.SearchLogger
	
	// External services
	dadataClient     EnhancedDaDataClient
	carAPIClient     EnhancedCarAPIClient
	repository       repositories.OfferRepository
	validator        validators.UserSearchRequestValidator
	
	// Configuration
	config           *config.SearchConfiguration
	
	// Legacy support
	legacyService    SearchService
}

// NewEnhancedSearchService creates a new enhanced search service
func NewEnhancedSearchService(
	config *config.SearchConfiguration,
	logger logging.SearchLogger,
	dadataClient EnhancedDaDataClient,
	carAPIClient EnhancedCarAPIClient,
	repository repositories.OfferRepository,
	validator validators.UserSearchRequestValidator,
	legacyService SearchService,
) EnhancedSearchService {
	
	// Initialize components
	queryProcessor := search.NewQueryProcessor()
	queryValidator := search.NewQueryValidator()
	resultRanker := search.NewResultRanker(logger)
	
	// Create strategy chain
	strategyChain := search.NewStrategyChain(logger)
	
	// Add strategies in order of priority
	exactMatchStrategy := search.NewExactMatchStrategy(carAPIClient, logger)
	strategyChain.AddStrategy(exactMatchStrategy)
	
	dadataAdapter := NewDaDataClientAdapter(dadataClient)
	dadataStrategy := search.NewDaDataStrategy(dadataAdapter, carAPIClient, logger)
	strategyChain.AddStrategy(dadataStrategy)
	
	fuzzyMatchStrategy := search.NewFuzzyMatchStrategy(carAPIClient, config, logger)
	strategyChain.AddStrategy(fuzzyMatchStrategy)
	
	return &enhancedSearchService{
		queryProcessor:  queryProcessor,
		queryValidator:  queryValidator,
		strategyChain:   strategyChain,
		resultRanker:    resultRanker,
		logger:          logger,
		dadataClient:    dadataClient,
		carAPIClient:    carAPIClient,
		repository:      repository,
		validator:       validator,
		config:          config,
		legacyService:   legacyService,
	}
}

// SearchTrims performs enhanced search with backward compatibility
func (ess *enhancedSearchService) SearchTrims(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error) {
	// Use enhanced search with default options
	options := SearchOptions{
		MaxResults:      10,
		EnableFallback:  true,
		IncludeDebugInfo: false,
		Timeout:         30 * time.Second,
		Strategy:        "auto",
	}
	
	enhancedResult, err := ess.SearchTrimsWithOptions(ctx, query, options)
	if err != nil {
		// Fallback to legacy service if enhanced search fails
		ess.logger.LogFallbackUsed("legacy_service", "enhanced search failed")
		return ess.legacyService.SearchTrims(ctx, query)
	}
	
	// Convert to legacy format
	return dtos.SearchTrimsResponse{
		Query: enhancedResult.Query,
		Count: enhancedResult.Count,
		Cars:  enhancedResult.Cars,
	}, nil
}

// SearchTrimsWithOptions performs enhanced search with detailed options
func (ess *enhancedSearchService) SearchTrimsWithOptions(ctx context.Context, query dtos.UserSearchQuery, options SearchOptions) (EnhancedSearchTrimsResponse, error) {
	startTime := time.Now()
	
	// Validate input
	if err := ess.validator.ValidateSearch(ctx, query); err != nil {
		return EnhancedSearchTrimsResponse{}, fmt.Errorf("validation failed: %w", err)
	}
	
	// Preprocess query
	preprocessed, err := ess.queryValidator.PreprocessQuery(query.Query)
	if err != nil {
		return EnhancedSearchTrimsResponse{}, fmt.Errorf("query preprocessing failed: %w", err)
	}
	
	// Parse query
	parsedQuery, err := ess.queryProcessor.ParseQuery(preprocessed.Sanitized)
	if err != nil {
		return EnhancedSearchTrimsResponse{}, fmt.Errorf("query parsing failed: %w", err)
	}
	
	// Create context with timeout
	searchCtx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	
	// Execute search strategy
	var searchResult *search.SearchResult
	var strategiesUsed []string
	
	if options.Strategy == "auto" {
		// Use strategy chain
		searchResult, err = ess.strategyChain.Execute(searchCtx, parsedQuery)
		if err != nil && options.EnableFallback {
			// Try fallback search
			searchResult, err = ess.performFallbackSearch(searchCtx, parsedQuery)
			if err == nil {
				strategiesUsed = append(strategiesUsed, "fallback")
			}
		}
	} else {
		// Use specific strategy
		searchResult, err = ess.executeSpecificStrategy(searchCtx, parsedQuery, options.Strategy)
		strategiesUsed = append(strategiesUsed, options.Strategy)
	}
	
	if err != nil {
		return EnhancedSearchTrimsResponse{}, fmt.Errorf("search execution failed: %w", err)
	}
	
	// Rank results
	rankedResults := ess.resultRanker.RankResults(parsedQuery, searchResult.Items)
	
	// Limit results
	if len(rankedResults) > options.MaxResults {
		rankedResults = rankedResults[:options.MaxResults]
	}
	
	processingTime := time.Since(startTime)
	
	// Build response
	response := EnhancedSearchTrimsResponse{
		Query: query.Query,
		Count: len(rankedResults),
		Cars:  rankedResults,
		Metadata: map[string]interface{}{
			"processing_time_ms": processingTime.Milliseconds(),
			"source_strategy":    searchResult.Source,
			"confidence":         searchResult.Confidence,
		},
	}
	
	// Add suggestions if no results or low confidence
	if len(rankedResults) == 0 || searchResult.Confidence < 0.5 {
		response.Suggestions = ess.generateSuggestions(parsedQuery, preprocessed)
	}
	
	// Add debug info if requested
	if options.IncludeDebugInfo {
		response.DebugInfo = &SearchDebugInfo{
			ParsedQuery:    parsedQuery,
			StrategiesUsed: strategiesUsed,
			ProcessingTime: processingTime,
			ValidationInfo: map[string]interface{}{
				"preprocessed": preprocessed,
				"is_valid":     preprocessed.IsValid,
			},
			RankingInfo: map[string]interface{}{
				"criteria": ess.resultRanker.GetCriteria(),
				"total_items_before_ranking": len(searchResult.Items),
			},
		}
	}
	
	ess.logger.LogSearchComplete(query.Query, len(rankedResults), processingTime)
	
	return response, nil
}

// SearchWithFallback performs search with comprehensive fallback mechanisms
func (ess *enhancedSearchService) SearchWithFallback(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error) {
	// Try enhanced search first
	result, err := ess.SearchTrims(ctx, query)
	if err == nil && len(result.Cars) > 0 {
		return result, nil
	}
	
	// Log fallback usage
	ess.logger.LogFallbackUsed("comprehensive_fallback", "primary search failed or returned no results")
	
	// Try different approaches
	fallbackApproaches := []func(context.Context, dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error){
		ess.trySimplifiedSearch,
		ess.tryKeywordSearch,
		ess.tryPopularModelsSearch,
		ess.legacyService.SearchTrims, // Ultimate fallback
	}
	
	for i, approach := range fallbackApproaches {
		result, err := approach(ctx, query)
		if err == nil && len(result.Cars) > 0 {
			ess.logger.LogFallbackUsed(fmt.Sprintf("fallback_approach_%d", i+1), "fallback succeeded")
			return result, nil
		}
	}
	
	// If all fallbacks fail, return empty result
	return dtos.SearchTrimsResponse{
		Query: query.Query,
		Count: 0,
		Cars:  []dtos.SearchTrimItem{},
	}, nil
}

// executeSpecificStrategy executes a specific search strategy
func (ess *enhancedSearchService) executeSpecificStrategy(ctx context.Context, query *search.ParsedQuery, strategy string) (*search.SearchResult, error) {
	switch strategy {
	case "exact":
		exactStrategy := search.NewExactMatchStrategy(ess.carAPIClient, ess.logger)
		if exactStrategy.CanHandle(query) {
			return exactStrategy.Search(ctx, query)
		}
		return nil, fmt.Errorf("exact match strategy cannot handle this query")
		
	case "dadata":
		dadataAdapter := NewDaDataClientAdapter(ess.dadataClient)
		dadataStrategy := search.NewDaDataStrategy(dadataAdapter, ess.carAPIClient, ess.logger)
		return dadataStrategy.Search(ctx, query)
		
	case "fuzzy":
		fuzzyStrategy := search.NewFuzzyMatchStrategy(ess.carAPIClient, ess.config, ess.logger)
		return fuzzyStrategy.Search(ctx, query)
		
	default:
		return nil, fmt.Errorf("unknown strategy: %s", strategy)
	}
}

// performFallbackSearch performs fallback search when primary strategies fail
func (ess *enhancedSearchService) performFallbackSearch(ctx context.Context, query *search.ParsedQuery) (*search.SearchResult, error) {
	ess.logger.LogFallbackUsed("fallback_search", "primary strategies failed")
	
	// Try fuzzy matching with lower thresholds
	fuzzyStrategy := search.NewFuzzyMatchStrategy(ess.carAPIClient, ess.config, ess.logger)
	
	// Lower the confidence requirement for fallback
	originalConfidence := query.Confidence
	query.Confidence = 0.3
	
	result, err := fuzzyStrategy.Search(ctx, query)
	
	// Restore original confidence
	query.Confidence = originalConfidence
	
	if err != nil || result == nil || len(result.Items) == 0 {
		// Try keyword-based search as last resort
		return ess.performKeywordFallback(ctx, query)
	}
	
	return result, nil
}

// performKeywordFallback performs keyword-based fallback search
func (ess *enhancedSearchService) performKeywordFallback(ctx context.Context, query *search.ParsedQuery) (*search.SearchResult, error) {
	// Extract individual words and try each as a model
	words := strings.Fields(strings.ToLower(query.Original))
	
	var allItems []dtos.SearchTrimItem
	
	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		
		response, err := ess.carAPIClient.GetTrimsByModel(ctx, word, 3)
		if err != nil || response == nil || len(response.Data) == 0 {
			continue
		}
		
		// Convert to search items
		for _, trim := range response.Data {
			item := dtos.SearchTrimItem{
				ID:          trim.ID,
				Year:        trim.Year,
				Make:        trim.Make,
				Model:       trim.Model,
				Trim:        trim.Trim,
				Description: trim.Description,
				MSRP:        trim.MSRP,
			}
			allItems = append(allItems, item)
		}
	}
	
	return &search.SearchResult{
		Items:      allItems,
		Source:     "keyword_fallback",
		Confidence: 0.2,
	}, nil
}

// Fallback search methods
func (ess *enhancedSearchService) trySimplifiedSearch(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error) {
	// Try with simplified query (remove extra words)
	simplified := ess.simplifyQuery(query.Query)
	if simplified != query.Query {
		simplifiedQuery := dtos.UserSearchQuery{
			Query:    simplified,
			Criteria: query.Criteria,
		}
		return ess.SearchTrims(ctx, simplifiedQuery)
	}
	return dtos.SearchTrimsResponse{}, fmt.Errorf("no simplification possible")
}

func (ess *enhancedSearchService) tryKeywordSearch(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error) {
	// Extract keywords and try each one
	keywords := strings.Fields(strings.ToLower(query.Query))
	
	for _, keyword := range keywords {
		if len(keyword) >= 3 {
			keywordQuery := dtos.UserSearchQuery{
				Query:    keyword,
				Criteria: query.Criteria,
			}
			
			result, err := ess.SearchTrims(ctx, keywordQuery)
			if err == nil && len(result.Cars) > 0 {
				return result, nil
			}
		}
	}
	
	return dtos.SearchTrimsResponse{}, fmt.Errorf("no keyword matches found")
}

func (ess *enhancedSearchService) tryPopularModelsSearch(ctx context.Context, query dtos.UserSearchQuery) (dtos.SearchTrimsResponse, error) {
	// Try popular models that might match the query
	queryLower := strings.ToLower(query.Query)
	
	for _, model := range ess.config.Fallback.PopularModels {
		if strings.Contains(queryLower, strings.ToLower(model)) {
			modelQuery := dtos.UserSearchQuery{
				Query:    model,
				Criteria: query.Criteria,
			}
			
			result, err := ess.SearchTrims(ctx, modelQuery)
			if err == nil && len(result.Cars) > 0 {
				return result, nil
			}
		}
	}
	
	return dtos.SearchTrimsResponse{}, fmt.Errorf("no popular model matches found")
}

// Helper methods
func (ess *enhancedSearchService) generateSuggestions(parsedQuery *search.ParsedQuery, preprocessed *search.PreprocessedQuery) []string {
	var suggestions []string
	
	// Add preprocessed suggestions
	suggestions = append(suggestions, preprocessed.Suggestions...)
	
	// Add popular model suggestions if query is vague
	if parsedQuery.Confidence < 0.5 {
		suggestions = append(suggestions, ess.config.Fallback.PopularModels[:5]...)
	}
	
	// Add make-specific suggestions
	if parsedQuery.Make != "" && parsedQuery.Model == "" {
		makeModels := map[string][]string{
			"toyota":     {"Camry", "Corolla", "Prius", "RAV4"},
			"honda":      {"Civic", "Accord", "CR-V", "Pilot"},
			"volkswagen": {"Golf", "Passat", "Jetta", "Tiguan"},
			"bmw":        {"X3", "X5", "3 Series", "5 Series"},
		}
		
		if models, exists := makeModels[strings.ToLower(parsedQuery.Make)]; exists {
			for _, model := range models {
				suggestions = append(suggestions, fmt.Sprintf("%s %s", parsedQuery.Make, model))
			}
		}
	}
	
	return ess.deduplicateSuggestions(suggestions)
}

func (ess *enhancedSearchService) simplifyQuery(query string) string {
	// Remove common words that might interfere with search
	stopWords := []string{"car", "auto", "vehicle", "the", "a", "an", "for", "sale"}
	
	words := strings.Fields(strings.ToLower(query))
	var filtered []string
	
	for _, word := range words {
		isStopWord := false
		for _, stopWord := range stopWords {
			if word == stopWord {
				isStopWord = true
				break
			}
		}
		if !isStopWord {
			filtered = append(filtered, word)
		}
	}
	
	return strings.Join(filtered, " ")
}

func (ess *enhancedSearchService) deduplicateSuggestions(suggestions []string) []string {
	seen := make(map[string]bool)
	var unique []string
	
	for _, suggestion := range suggestions {
		if !seen[suggestion] && suggestion != "" {
			seen[suggestion] = true
			unique = append(unique, suggestion)
		}
	}
	
	// Limit to reasonable number
	if len(unique) > 10 {
		unique = unique[:10]
	}
	
	return unique
}