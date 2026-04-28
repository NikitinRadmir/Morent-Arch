package config

import (
	"encoding/json"
	"os"
	"time"
)

type SearchConfiguration struct {
	DaData struct {
		Timeout     time.Duration `json:"timeout"`
		MaxResults  int          `json:"max_results"`
		Enabled     bool         `json:"enabled"`
	} `json:"dadata"`
	
	CarAPI struct {
		Timeout     time.Duration `json:"timeout"`
		MaxRetries  int          `json:"max_retries"`
		RateLimit   int          `json:"rate_limit"`
	} `json:"carapi"`
	
	Fallback struct {
		Enabled           bool     `json:"enabled"`
		FuzzyThreshold    float64  `json:"fuzzy_threshold"`
		PopularModels     []string `json:"popular_models"`
	} `json:"fallback"`
	
	Logging struct {
		Level           string `json:"level"`
		EnableDebugInfo bool   `json:"enable_debug_info"`
	} `json:"logging"`
}

func LoadSearchConfig() (*SearchConfiguration, error) {
	config := &SearchConfiguration{}
	
	// Set defaults
	config.DaData.Timeout = 10 * time.Second
	config.DaData.MaxResults = 5
	config.DaData.Enabled = true
	
	config.CarAPI.Timeout = 15 * time.Second
	config.CarAPI.MaxRetries = 3
	config.CarAPI.RateLimit = 100
	
	config.Fallback.Enabled = true
	config.Fallback.FuzzyThreshold = 0.7
	config.Fallback.PopularModels = []string{"Golf", "Camry", "Civic", "Corolla", "Focus"}
	
	config.Logging.Level = "info"
	config.Logging.EnableDebugInfo = false
	
	// Try to load from file
	if configFile := os.Getenv("SEARCH_CONFIG_FILE"); configFile != "" {
		if data, err := os.ReadFile(configFile); err == nil {
			if err := json.Unmarshal(data, config); err != nil {
				return nil, err
			}
		}
	}
	
	// Override with environment variables
	if timeout := os.Getenv("DADATA_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.DaData.Timeout = d
		}
	}
	
	if timeout := os.Getenv("CARAPI_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			config.CarAPI.Timeout = d
		}
	}
	
	if level := os.Getenv("SEARCH_LOG_LEVEL"); level != "" {
		config.Logging.Level = level
	}
	
	return config, nil
}

func (c *SearchConfiguration) Validate() error {
	if c.DaData.Timeout <= 0 {
		c.DaData.Timeout = 10 * time.Second
	}
	if c.CarAPI.Timeout <= 0 {
		c.CarAPI.Timeout = 15 * time.Second
	}
	if c.CarAPI.MaxRetries <= 0 {
		c.CarAPI.MaxRetries = 3
	}
	if c.Fallback.FuzzyThreshold <= 0 || c.Fallback.FuzzyThreshold > 1 {
		c.Fallback.FuzzyThreshold = 0.7
	}
	return nil
}