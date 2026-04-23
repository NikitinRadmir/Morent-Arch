package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DadataService struct {
	apiKey    string
	secretKey string
	client    *http.Client
	baseURL   string
}

type DadataRequest struct {
	Query string `json:"query"`
}

type DadataResponse struct {
	Suggestions []DadataSuggestion `json:"suggestions"`
}

type DadataSuggestion struct {
	Value string      `json:"value"`
	Data  DadataData  `json:"data"`
}

type DadataData struct {
	Brand string `json:"brand"`
	Model string `json:"model"`
}

func NewDadataService(apiKey, secretKey string) *DadataService {
	return &DadataService{
		apiKey:    apiKey,
		secretKey: secretKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/car",
	}
}

func (s *DadataService) NormalizeCarName(query string) (string, error) {
	if s.apiKey == "" || s.secretKey == "" {
		return query, nil 
	}

	reqBody := DadataRequest{Query: query}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Token "+s.apiKey)
	req.Header.Set("X-Secret", s.secretKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return query, nil 
	}

	var dadataResp DadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&dadataResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(dadataResp.Suggestions) > 0 {
		suggestion := dadataResp.Suggestions[0]
		if suggestion.Data.Brand != "" && suggestion.Data.Model != "" {
			return fmt.Sprintf("%s %s", suggestion.Data.Brand, suggestion.Data.Model), nil
		}
		return suggestion.Value, nil
	}

	return query, nil
}