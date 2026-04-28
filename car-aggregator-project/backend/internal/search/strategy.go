package search

import (
	"context"
	"fmt"
	"sort"
	"time"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// SearchStrategy defines the interface for different search strategies
type SearchStrategy interface {
	Search(ctx context.Context, query *ParsedQuery) (*SearchResult, error)
	CanHandle(query *ParsedQuery) bool
	Priority() int
	Name() string
}

// SearchResult represents the result from a search strategy
type SearchResult struct {
	Items      []dtos.SearchTrimItem `json:"items"`
	Source     string                `json:"source"`
	Confidence float64               `json:"confidence"`
	Duration   time.Duration         `json:"duration"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// StrategyChain manages the execution of multiple search strategies
type StrategyChain struct {
	strategies []SearchStrategy
	logger     logging.SearchLogger
}

// NewStrategyChain creates a new strategy chain
func NewStrategyChain(logger logging.SearchLogger) *StrategyChain {
	return &StrategyChain{
		strategies: make([]SearchStrategy, 0),
		logger:     logger,
	}
}

// AddStrategy adds a strategy to the chain
func (sc *StrategyChain) AddStrategy(strategy SearchStrategy) {
	sc.strategies = append(sc.strategies, strategy)
	
	// Sort strategies by priority (higher priority first)
	sort.Slice(sc.strategies, func(i, j int) bool {
		return sc.strategies[i].Priority() > sc.strategies[j].Priority()
	})
}

// Execute runs the strategy chain and returns the best result
func (sc *StrategyChain) Execute(ctx context.Context, query *ParsedQuery) (*SearchResult, error) {
	sc.logger.LogSearchStart(query.Original)
	startTime := time.Now()
	
	var bestResult *SearchResult
	var lastError error
	
	for _, strategy := range sc.strategies {
		if !strategy.CanHandle(query) {
			continue
		}
		
		sc.logger.LogFallbackUsed(strategy.Name(), "executing strategy")
		
		result, err := strategy.Search(ctx, query)
		if err != nil {
			lastError = err
			sc.logger.LogError("strategy_failed", err, map[string]interface{}{
				"strategy": strategy.Name(),
				"query":    query.Original,
			})
			continue
		}
		
		if result != nil && len(result.Items) > 0 {
			if bestResult == nil || sc.isBetterResult(result, bestResult) {
				bestResult = result
			}
			
			// If we have a high-confidence result, we can stop here
			if result.Confidence >= 0.9 {
				break
			}
		}
	}
	
	duration := time.Since(startTime)
	
	if bestResult == nil {
		if lastError != nil {
			return nil, fmt.Errorf("all strategies failed, last error: %w", lastError)
		}
		return nil, fmt.Errorf("no results found for query: %s", query.Original)
	}
	
	sc.logger.LogSearchComplete(query.Original, len(bestResult.Items), duration)
	return bestResult, nil
}

// ExecuteAll runs all applicable strategies and returns combined results
func (sc *StrategyChain) ExecuteAll(ctx context.Context, query *ParsedQuery) ([]*SearchResult, error) {
	sc.logger.LogSearchStart(query.Original)
	
	var results []*SearchResult
	
	for _, strategy := range sc.strategies {
		if !strategy.CanHandle(query) {
			continue
		}
		
		result, err := strategy.Search(ctx, query)
		if err != nil {
			sc.logger.LogError("strategy_failed", err, map[string]interface{}{
				"strategy": strategy.Name(),
				"query":    query.Original,
			})
			continue
		}
		
		if result != nil && len(result.Items) > 0 {
			results = append(results, result)
		}
	}
	
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for query: %s", query.Original)
	}
	
	return results, nil
}

// isBetterResult determines if result1 is better than result2
func (sc *StrategyChain) isBetterResult(result1, result2 *SearchResult) bool {
	// Higher confidence is better
	if result1.Confidence > result2.Confidence {
		return true
	}
	
	// If confidence is similar, more results is better
	if absFloat(result1.Confidence-result2.Confidence) < 0.1 {
		return len(result1.Items) > len(result2.Items)
	}
	
	return false
}

// BaseStrategy provides common functionality for all strategies
type BaseStrategy struct {
	name     string
	priority int
	logger   logging.SearchLogger
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(name string, priority int, logger logging.SearchLogger) *BaseStrategy {
	return &BaseStrategy{
		name:     name,
		priority: priority,
		logger:   logger,
	}
}

// Name returns the strategy name
func (bs *BaseStrategy) Name() string {
	return bs.name
}

// Priority returns the strategy priority
func (bs *BaseStrategy) Priority() int {
	return bs.priority
}

// LogStart logs the start of strategy execution
func (bs *BaseStrategy) LogStart(query *ParsedQuery) {
	bs.logger.LogFallbackUsed(bs.name, fmt.Sprintf("starting search for: %s", query.Original))
}

// LogResult logs the result of strategy execution
func (bs *BaseStrategy) LogResult(query *ParsedQuery, result *SearchResult, err error) {
	if err != nil {
		bs.logger.LogError("strategy_error", err, map[string]interface{}{
			"strategy": bs.name,
			"query":    query.Original,
		})
	} else if result != nil {
		bs.logger.LogSearchComplete(query.Original, len(result.Items), result.Duration)
	}
}

// SearchOptions provides configuration for search strategies
type SearchOptions struct {
	MaxResults     int           `json:"max_results"`
	MinConfidence  float64       `json:"min_confidence"`
	Timeout        time.Duration `json:"timeout"`
	EnableFallback bool          `json:"enable_fallback"`
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		MaxResults:     10,
		MinConfidence:  0.3,
		Timeout:        30 * time.Second,
		EnableFallback: true,
	}
}

// StrategyResult represents the result from a single strategy execution
type StrategyResult struct {
	Strategy   string                `json:"strategy"`
	Success    bool                  `json:"success"`
	Items      []dtos.SearchTrimItem `json:"items"`
	Confidence float64               `json:"confidence"`
	Duration   time.Duration         `json:"duration"`
	Error      string                `json:"error,omitempty"`
}

// CombinedSearchResult represents results from multiple strategies
type CombinedSearchResult struct {
	Query           string            `json:"query"`
	TotalResults    int               `json:"total_results"`
	BestStrategy    string            `json:"best_strategy"`
	Items           []dtos.SearchTrimItem `json:"items"`
	StrategyResults []StrategyResult  `json:"strategy_results"`
	Duration        time.Duration     `json:"duration"`
}

// ResultCombiner combines results from multiple strategies
type ResultCombiner struct {
	logger logging.SearchLogger
}

// NewResultCombiner creates a new result combiner
func NewResultCombiner(logger logging.SearchLogger) *ResultCombiner {
	return &ResultCombiner{
		logger: logger,
	}
}

// CombineResults combines multiple search results into one
func (rc *ResultCombiner) CombineResults(results []*SearchResult) *SearchResult {
	if len(results) == 0 {
		return nil
	}
	
	if len(results) == 1 {
		return results[0]
	}
	
	// Combine all items
	var allItems []dtos.SearchTrimItem
	var totalConfidence float64
	var bestSource string
	var maxConfidence float64
	
	for _, result := range results {
		allItems = append(allItems, result.Items...)
		totalConfidence += result.Confidence
		
		if result.Confidence > maxConfidence {
			maxConfidence = result.Confidence
			bestSource = result.Source
		}
	}
	
	// Remove duplicates
	uniqueItems := rc.removeDuplicates(allItems)
	
	// Calculate average confidence
	avgConfidence := totalConfidence / float64(len(results))
	
	return &SearchResult{
		Items:      uniqueItems,
		Source:     bestSource,
		Confidence: avgConfidence,
		Metadata: map[string]interface{}{
			"combined_from": len(results),
			"max_confidence": maxConfidence,
		},
	}
}

// removeDuplicates removes duplicate items based on make, model, and year
func (rc *ResultCombiner) removeDuplicates(items []dtos.SearchTrimItem) []dtos.SearchTrimItem {
	seen := make(map[string]bool)
	var unique []dtos.SearchTrimItem
	
	for _, item := range items {
		key := fmt.Sprintf("%s_%s_%d", item.Make, item.Model, item.Year)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, item)
		}
	}
	
	return unique
}

// Helper function to calculate absolute difference
func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// StrategyFactory creates different types of strategies
type StrategyFactory struct {
	logger logging.SearchLogger
}

// NewStrategyFactory creates a new strategy factory
func NewStrategyFactory(logger logging.SearchLogger) *StrategyFactory {
	return &StrategyFactory{
		logger: logger,
	}
}

// CreateDefaultChain creates a default strategy chain with all strategies
func (sf *StrategyFactory) CreateDefaultChain() *StrategyChain {
	chain := NewStrategyChain(sf.logger)
	
	// Add strategies in order of preference
	// (actual strategy implementations will be added in subsequent tasks)
	
	return chain
}

// Strategy priority constants
const (
	PriorityExactMatch = 100
	PriorityDaData     = 80
	PriorityFuzzyMatch = 60
	PriorityFallback   = 40
	PriorityDefault    = 20
)

// Strategy names
const (
	StrategyNameExactMatch = "exact_match"
	StrategyNameDaData     = "dadata"
	StrategyNameFuzzyMatch = "fuzzy_match"
	StrategyNameFallback   = "fallback"
	StrategyNameDefault    = "default"
)