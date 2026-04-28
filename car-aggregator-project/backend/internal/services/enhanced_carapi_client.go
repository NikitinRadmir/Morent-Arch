package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"car-aggregator/internal/config"
	"car-aggregator/internal/dtos"
	"car-aggregator/internal/logging"
)

type EnhancedCarAPIClient interface {
	Login(ctx context.Context) error
	GetTrimsByMakeAndModel(ctx context.Context, make, model string, limit int) (*dtos.TrimsResponse, error)
	GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error)
	SearchTrims(ctx context.Context, searchTerms []string, limit int) (*dtos.TrimsResponse, error)
	GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error)
}

type enhancedCarAPIClient struct {
	baseURL       string
	apiToken      string
	apiSecret     string
	jwtToken      string
	httpClient    *http.Client
	config        *config.SearchConfiguration
	logger        logging.SearchLogger
	cache         *CarAPICache
	rateLimiter   *RateLimiter
	circuitBreaker *CircuitBreaker
	mutex         sync.RWMutex
}

type CarAPICache struct {
	data   map[string]*CacheEntry
	mutex  sync.RWMutex
	ttl    time.Duration
}

type CacheEntry struct {
	data      interface{}
	timestamp time.Time
}

type RateLimiter struct {
	tokens    int
	maxTokens int
	refillRate time.Duration
	lastRefill time.Time
	mutex     sync.Mutex
}

func NewEnhancedCarAPIClient(apiToken, apiSecret string, config *config.SearchConfiguration, logger logging.SearchLogger) EnhancedCarAPIClient {
	return &enhancedCarAPIClient{
		baseURL:   "https://carapi.app/api",
		apiToken:  apiToken,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Timeout: config.CarAPI.Timeout,
		},
		config: config,
		logger: logger,
		cache: &CarAPICache{
			data: make(map[string]*CacheEntry),
			ttl:  15 * time.Minute,
		},
		rateLimiter: &RateLimiter{
			tokens:     config.CarAPI.RateLimit,
			maxTokens:  config.CarAPI.RateLimit,
			refillRate: time.Minute,
			lastRefill: time.Now(),
		},
		circuitBreaker: &CircuitBreaker{
			timeout:     30 * time.Second,
			maxFailures: 5,
			state:       CircuitClosed,
		},
	}
}

func (c *enhancedCarAPIClient) Login(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.jwtToken != "" {
		// Check if token is still valid (simple check)
		return nil
	}
	
	loginReq := dtos.LoginRequest{
		APIToken:  c.apiToken,
		APISecret: c.apiSecret,
	}
	
	body, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/auth/login", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}
	
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	token := buf.String()
	if token == "" {
		return fmt.Errorf("received empty token")
	}
	
	c.jwtToken = token
	return nil
}

func (c *enhancedCarAPIClient) GetTrimsByMakeAndModel(ctx context.Context, make, model string, limit int) (*dtos.TrimsResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("trims_make_model_%s_%s_%d", strings.ToLower(make), strings.ToLower(model), limit)
	if cached := c.cache.get(cacheKey); cached != nil {
		if response, ok := cached.(*dtos.TrimsResponse); ok {
			c.logger.LogCarAPIResponse(len(response.Data), 0) // 0 duration for cached response
			return response, nil
		}
	}
	
	// Check circuit breaker
	if !c.circuitBreaker.canExecute() {
		return nil, fmt.Errorf("CarAPI service is temporarily unavailable")
	}
	
	// Check rate limiter
	if !c.rateLimiter.allowRequest() {
		return nil, fmt.Errorf("rate limit exceeded for CarAPI")
	}
	
	startTime := time.Now()
	c.logger.LogCarAPIRequest(make, model, limit)
	
	// Ensure we're authenticated
	if err := c.Login(ctx); err != nil {
		c.circuitBreaker.recordFailure()
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	
	// Build URL with both make and model parameters
	apiURL := fmt.Sprintf("%s/trims/v2?limit=%d&make=%s&model=%s", 
		c.baseURL, limit, url.QueryEscape(make), url.QueryEscape(model))
	
	response, err := c.executeRequest(ctx, apiURL)
	if err != nil {
		c.circuitBreaker.recordFailure()
		return nil, err
	}
	
	// Cache the response
	c.cache.set(cacheKey, response)
	
	c.circuitBreaker.recordSuccess()
	duration := time.Since(startTime)
	c.logger.LogCarAPIResponse(len(response.Data), duration)
	
	return response, nil
}

func (c *enhancedCarAPIClient) GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("trims_model_%s_%d", strings.ToLower(model), limit)
	if cached := c.cache.get(cacheKey); cached != nil {
		if response, ok := cached.(*dtos.TrimsResponse); ok {
			c.logger.LogCarAPIResponse(len(response.Data), 0)
			return response, nil
		}
	}
	
	// Check circuit breaker and rate limiter
	if !c.circuitBreaker.canExecute() {
		return nil, fmt.Errorf("CarAPI service is temporarily unavailable")
	}
	
	if !c.rateLimiter.allowRequest() {
		return nil, fmt.Errorf("rate limit exceeded for CarAPI")
	}
	
	startTime := time.Now()
	c.logger.LogCarAPIRequest("", model, limit)
	
	// Ensure we're authenticated
	if err := c.Login(ctx); err != nil {
		c.circuitBreaker.recordFailure()
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	
	// Build URL with model parameter only
	apiURL := fmt.Sprintf("%s/trims/v2?limit=%d&model=%s", 
		c.baseURL, limit, url.QueryEscape(model))
	
	response, err := c.executeRequest(ctx, apiURL)
	if err != nil {
		c.circuitBreaker.recordFailure()
		return nil, err
	}
	
	// Cache the response
	c.cache.set(cacheKey, response)
	
	c.circuitBreaker.recordSuccess()
	duration := time.Since(startTime)
	c.logger.LogCarAPIResponse(len(response.Data), duration)
	
	return response, nil
}

func (c *enhancedCarAPIClient) SearchTrims(ctx context.Context, searchTerms []string, limit int) (*dtos.TrimsResponse, error) {
	// Try different combinations of search terms
	var bestResponse *dtos.TrimsResponse
	var bestScore int
	
	for _, term := range searchTerms {
		response, err := c.GetTrimsByModel(ctx, term, limit)
		if err != nil {
			continue
		}
		
		if response != nil && len(response.Data) > bestScore {
			bestResponse = response
			bestScore = len(response.Data)
		}
	}
	
	if bestResponse == nil {
		return nil, fmt.Errorf("no results found for search terms: %v", searchTerms)
	}
	
	return bestResponse, nil
}

func (c *enhancedCarAPIClient) GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("specs_%d_%s_%s", year, strings.ToLower(make), strings.ToLower(model))
	if cached := c.cache.get(cacheKey); cached != nil {
		if specs, ok := cached.(VehicleSpecs); ok {
			return specs.Seats, specs.Fuel, nil
		}
	}
	
	seats, err := c.getSeats(ctx, year, make, model)
	if err != nil {
		seats = 0 // Default value
	}
	
	fuel, err := c.getFuelConsumption(ctx, year, make, model)
	if err != nil {
		fuel = 0 // Default value
	}
	
	// Cache the specs
	specs := VehicleSpecs{Seats: seats, Fuel: fuel}
	c.cache.set(cacheKey, specs)
	
	return seats, fuel, nil
}

type VehicleSpecs struct {
	Seats int
	Fuel  float64
}

func (c *enhancedCarAPIClient) executeRequest(ctx context.Context, apiURL string) (*dtos.TrimsResponse, error) {
	var lastErr error
	
	// Retry logic with exponential backoff
	for attempt := 0; attempt < c.config.CarAPI.MaxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.jwtToken)
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.logger.LogError("carapi_request_failed", err, map[string]interface{}{
				"url":     apiURL,
				"attempt": attempt + 1,
			})
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusUnauthorized {
			// Token might be expired, try to re-authenticate
			c.mutex.Lock()
			c.jwtToken = ""
			c.mutex.Unlock()
			
			if err := c.Login(ctx); err != nil {
				lastErr = err
				continue
			}
			
			// Retry the request with new token
			req.Header.Set("Authorization", "Bearer "+c.jwtToken)
			resp, err = c.httpClient.Do(req)
			if err != nil {
				lastErr = err
				continue
			}
		}
		
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("request failed with status: %d", resp.StatusCode)
			continue
		}
		
		var trimsResp dtos.TrimsResponse
		if err := json.NewDecoder(resp.Body).Decode(&trimsResp); err != nil {
			lastErr = fmt.Errorf("failed to decode response: %w", err)
			continue
		}
		
		return &trimsResp, nil
	}
	
	return nil, fmt.Errorf("CarAPI request failed after %d retries: %w", c.config.CarAPI.MaxRetries, lastErr)
}

func (c *enhancedCarAPIClient) getSeats(ctx context.Context, year int, make, model string) (int, error) {
	endpoint := fmt.Sprintf(
		"%s/bodies/v2?limit=1&year=%d&make=%s&model=%s",
		c.baseURL,
		year,
		url.QueryEscape(make),
		url.QueryEscape(model),
	)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create bodies request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.jwtToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute bodies request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bodies request failed with status: %d", resp.StatusCode)
	}
	
	var out dtos.BodiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("failed to decode bodies response: %w", err)
	}
	if len(out.Data) == 0 {
		return 0, nil
	}
	return out.Data[0].Seats, nil
}

func (c *enhancedCarAPIClient) getFuelConsumption(ctx context.Context, year int, make, model string) (float64, error) {
	endpoint := fmt.Sprintf(
		"%s/engines/v2?limit=1&year=%d&make=%s&model=%s",
		c.baseURL,
		year,
		url.QueryEscape(make),
		url.QueryEscape(model),
	)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create engines request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.jwtToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to execute engines request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("engines request failed with status: %d", resp.StatusCode)
	}
	
	var out dtos.EnginesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("failed to decode engines response: %w", err)
	}
	if len(out.Data) == 0 {
		return 0, nil
	}
	
	engine := out.Data[0]
	size, _ := strconv.ParseFloat(strings.TrimSpace(engine.Size), 64)
	fuelType := strings.ToLower(engine.FuelType)
	engineType := strings.ToLower(engine.EngineType)
	
	// Improved fuel consumption estimation
	estimated := 5.8 + size*1.3
	if strings.Contains(engine.Cylinders, "6") {
		estimated += 1.4
	}
	if strings.Contains(engine.Cylinders, "8") {
		estimated += 2.2
	}
	if strings.Contains(engineType, "hybrid") {
		estimated -= 2.0
	}
	if strings.Contains(fuelType, "diesel") {
		estimated -= 0.6
	}
	
	if estimated < 3.8 {
		estimated = 3.8
	}
	if estimated > 18 {
		estimated = 18
	}
	
	return estimated, nil
}

// Cache implementation
func (cache *CarAPICache) get(key string) interface{} {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	
	entry, exists := cache.data[key]
	if !exists {
		return nil
	}
	
	if time.Since(entry.timestamp) > cache.ttl {
		delete(cache.data, key)
		return nil
	}
	
	return entry.data
}

func (cache *CarAPICache) set(key string, data interface{}) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	
	cache.data[key] = &CacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

// Rate limiter implementation
func (rl *RateLimiter) allowRequest() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	if now.Sub(rl.lastRefill) >= rl.refillRate {
		rl.tokens = rl.maxTokens
		rl.lastRefill = now
	}
	
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	
	return false
}