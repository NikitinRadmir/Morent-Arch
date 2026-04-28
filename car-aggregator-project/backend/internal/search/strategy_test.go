package search

import (
	"context"
	"fmt"
	"testing"
	"time"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

// Mock strategy for testing
type mockStrategy struct {
	*BaseStrategy
	canHandle    bool
	shouldFail   bool
	resultItems  []dtos.SearchTrimItem
	confidence   float64
}

func newMockStrategy(name string, priority int, canHandle bool, shouldFail bool, items []dtos.SearchTrimItem, confidence float64) *mockStrategy {
	logger := logging.NewSearchLogger("debug", true)
	return &mockStrategy{
		BaseStrategy: NewBaseStrategy(name, priority, logger),
		canHandle:    canHandle,
		shouldFail:   shouldFail,
		resultItems:  items,
		confidence:   confidence,
	}
}

func (ms *mockStrategy) Search(ctx context.Context, query *ParsedQuery) (*SearchResult, error) {
	if ms.shouldFail {
		return nil, fmt.Errorf("mock strategy failed")
	}
	
	return &SearchResult{
		Items:      ms.resultItems,
		Source:     ms.name,
		Confidence: ms.confidence,
		Duration:   10 * time.Millisecond,
	}, nil
}

func (ms *mockStrategy) CanHandle(query *ParsedQuery) bool {
	return ms.canHandle
}

func TestStrategyChain_AddStrategy(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	chain := NewStrategyChain(logger)
	
	// Add strategies with different priorities
	strategy1 := newMockStrategy("low_priority", 10, true, false, nil, 0.5)
	strategy2 := newMockStrategy("high_priority", 90, true, false, nil, 0.8)
	strategy3 := newMockStrategy("medium_priority", 50, true, false, nil, 0.6)
	
	chain.AddStrategy(strategy1)
	chain.AddStrategy(strategy2)
	chain.AddStrategy(strategy3)
	
	// Check that strategies are sorted by priority (highest first)
	if len(chain.strategies) != 3 {
		t.Errorf("Expected 3 strategies, got %d", len(chain.strategies))
	}
	
	if chain.strategies[0].Priority() != 90 {
		t.Errorf("Expected highest priority strategy first, got priority %d", chain.strategies[0].Priority())
	}
	
	if chain.strategies[1].Priority() != 50 {
		t.Errorf("Expected medium priority strategy second, got priority %d", chain.strategies[1].Priority())
	}
	
	if chain.strategies[2].Priority() != 10 {
		t.Errorf("Expected lowest priority strategy last, got priority %d", chain.strategies[2].Priority())
	}
}

func TestStrategyChain_Execute(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	chain := NewStrategyChain(logger)
	
	// Create test items
	items1 := []dtos.SearchTrimItem{
		{Make: "Toyota", Model: "Camry", Year: 2020},
	}
	items2 := []dtos.SearchTrimItem{
		{Make: "Honda", Model: "Civic", Year: 2021},
		{Make: "Honda", Model: "Accord", Year: 2021},
	}
	
	// Add strategies
	strategy1 := newMockStrategy("strategy1", 80, true, false, items1, 0.7)
	strategy2 := newMockStrategy("strategy2", 90, true, false, items2, 0.9)
	strategy3 := newMockStrategy("strategy3", 70, false, false, nil, 0.5) // Can't handle
	
	chain.AddStrategy(strategy1)
	chain.AddStrategy(strategy2)
	chain.AddStrategy(strategy3)
	
	query := &ParsedQuery{
		Original: "test query",
		Make:     "Honda",
		Model:    "Civic",
	}
	
	result, err := chain.Execute(context.Background(), query)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	
	if result == nil {
		t.Error("Execute() returned nil result")
		return
	}
	
	// Should return the result from strategy2 (highest confidence)
	if result.Source != "strategy2" {
		t.Errorf("Expected result from strategy2, got %s", result.Source)
	}
	
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
	
	if result.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", result.Confidence)
	}
}

func TestStrategyChain_ExecuteWithFailures(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	chain := NewStrategyChain(logger)
	
	items := []dtos.SearchTrimItem{
		{Make: "Toyota", Model: "Camry", Year: 2020},
	}
	
	// Add strategies where first ones fail
	strategy1 := newMockStrategy("failing_strategy1", 90, true, true, nil, 0.0)
	strategy2 := newMockStrategy("failing_strategy2", 80, true, true, nil, 0.0)
	strategy3 := newMockStrategy("working_strategy", 70, true, false, items, 0.6)
	
	chain.AddStrategy(strategy1)
	chain.AddStrategy(strategy2)
	chain.AddStrategy(strategy3)
	
	query := &ParsedQuery{
		Original: "test query",
		Make:     "Toyota",
		Model:    "Camry",
	}
	
	result, err := chain.Execute(context.Background(), query)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	
	if result == nil {
		t.Error("Execute() returned nil result")
		return
	}
	
	// Should return the result from the working strategy
	if result.Source != "working_strategy" {
		t.Errorf("Expected result from working_strategy, got %s", result.Source)
	}
}

func TestStrategyChain_ExecuteAllFail(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	chain := NewStrategyChain(logger)
	
	// Add only failing strategies
	strategy1 := newMockStrategy("failing_strategy1", 90, true, true, nil, 0.0)
	strategy2 := newMockStrategy("failing_strategy2", 80, true, true, nil, 0.0)
	
	chain.AddStrategy(strategy1)
	chain.AddStrategy(strategy2)
	
	query := &ParsedQuery{
		Original: "test query",
		Make:     "Toyota",
		Model:    "Camry",
	}
	
	result, err := chain.Execute(context.Background(), query)
	if err == nil {
		t.Error("Execute() should have returned an error when all strategies fail")
	}
	
	if result != nil {
		t.Error("Execute() should have returned nil result when all strategies fail")
	}
}

func TestStrategyChain_ExecuteAll(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	chain := NewStrategyChain(logger)
	
	items1 := []dtos.SearchTrimItem{
		{Make: "Toyota", Model: "Camry", Year: 2020},
	}
	items2 := []dtos.SearchTrimItem{
		{Make: "Honda", Model: "Civic", Year: 2021},
	}
	
	strategy1 := newMockStrategy("strategy1", 80, true, false, items1, 0.7)
	strategy2 := newMockStrategy("strategy2", 90, true, false, items2, 0.9)
	strategy3 := newMockStrategy("strategy3", 70, true, true, nil, 0.0) // Fails
	
	chain.AddStrategy(strategy1)
	chain.AddStrategy(strategy2)
	chain.AddStrategy(strategy3)
	
	query := &ParsedQuery{
		Original: "test query",
		Make:     "Honda",
		Model:    "Civic",
	}
	
	results, err := chain.ExecuteAll(context.Background(), query)
	if err != nil {
		t.Errorf("ExecuteAll() error = %v", err)
		return
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	
	// Results should be from strategy1 and strategy2 (strategy3 fails)
	sources := make(map[string]bool)
	for _, result := range results {
		sources[result.Source] = true
	}
	
	if !sources["strategy1"] || !sources["strategy2"] {
		t.Error("ExecuteAll() should return results from both working strategies")
	}
}

func TestResultCombiner_CombineResults(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	combiner := NewResultCombiner(logger)
	
	items1 := []dtos.SearchTrimItem{
		{Make: "Toyota", Model: "Camry", Year: 2020, ID: 1},
	}
	items2 := []dtos.SearchTrimItem{
		{Make: "Honda", Model: "Civic", Year: 2021, ID: 2},
		{Make: "Toyota", Model: "Camry", Year: 2020, ID: 1}, // Duplicate
	}
	
	result1 := &SearchResult{
		Items:      items1,
		Source:     "source1",
		Confidence: 0.7,
	}
	
	result2 := &SearchResult{
		Items:      items2,
		Source:     "source2",
		Confidence: 0.9,
	}
	
	combined := combiner.CombineResults([]*SearchResult{result1, result2})
	
	if combined == nil {
		t.Error("CombineResults() returned nil")
		return
	}
	
	// Should have 2 unique items (duplicate removed)
	if len(combined.Items) != 2 {
		t.Errorf("Expected 2 unique items, got %d", len(combined.Items))
	}
	
	// Should use the source with highest confidence
	if combined.Source != "source2" {
		t.Errorf("Expected source2 (highest confidence), got %s", combined.Source)
	}
	
	// Should calculate average confidence
	expectedConfidence := (0.7 + 0.9) / 2
	if combined.Confidence != expectedConfidence {
		t.Errorf("Expected confidence %f, got %f", expectedConfidence, combined.Confidence)
	}
}

func TestResultCombiner_RemoveDuplicates(t *testing.T) {
	logger := logging.NewSearchLogger("debug", true)
	combiner := NewResultCombiner(logger)
	
	items := []dtos.SearchTrimItem{
		{Make: "Toyota", Model: "Camry", Year: 2020, ID: 1},
		{Make: "Honda", Model: "Civic", Year: 2021, ID: 2},
		{Make: "Toyota", Model: "Camry", Year: 2020, ID: 3}, // Duplicate by make/model/year
		{Make: "Toyota", Model: "Camry", Year: 2021, ID: 4}, // Different year, not duplicate
	}
	
	unique := combiner.removeDuplicates(items)
	
	if len(unique) != 3 {
		t.Errorf("Expected 3 unique items, got %d", len(unique))
	}
	
	// Check that we have the expected combinations
	seen := make(map[string]bool)
	for _, item := range unique {
		key := fmt.Sprintf("%s_%s_%d", item.Make, item.Model, item.Year)
		if seen[key] {
			t.Errorf("Found duplicate key: %s", key)
		}
		seen[key] = true
	}
}