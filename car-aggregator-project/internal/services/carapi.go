package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CarAPIService struct {
	token   string
	secret  string
	client  *http.Client
	baseURL string
}

type CarAPIResponse struct {
	Data []CarData `json:"data"`
}

type CarData struct {
	ID    int    `json:"id"`
	Make  string `json:"make"`
	Model string `json:"model"`
	Year  int    `json:"year"`
	Trims []Trim `json:"trims,omitempty"`
}

type Trim struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MSRP        int    `json:"msrp"`
	Invoice     int    `json:"invoice"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
}

type TrimsResponse struct {
	Data []Trim `json:"data"`
}

func NewCarAPIService(token, secret string) *CarAPIService {
	return &CarAPIService{
		token:  token,
		secret: secret,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL: "https://carapi.app/api",
	}
}

func (s *CarAPIService) SearchCars(query string) ([]CarData, error) {
	if s.token == "" {
		return nil, fmt.Errorf("CarAPI token not configured")
	}

	parts := strings.Fields(strings.TrimSpace(query))
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty search query")
	}

	u, err := url.Parse(s.baseURL + "/cars")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	params := url.Values{}
	if len(parts) >= 1 {
		params.Add("make", parts[0])
	}
	if len(parts) >= 2 {
		params.Add("model", strings.Join(parts[1:], " "))
	}
	params.Add("limit", "50")
	u.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Partner-Token", s.secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var apiResp CarAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apiResp.Data, nil
}

func (s *CarAPIService) GetTrims(carID int) ([]Trim, error) {
	if s.token == "" {
		return nil, fmt.Errorf("CarAPI token not configured")
	}

	url := fmt.Sprintf("%s/trims?car_id=%d", s.baseURL, carID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Partner-Token", s.secret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var trimsResp TrimsResponse
	if err := json.NewDecoder(resp.Body).Decode(&trimsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return trimsResp.Data, nil
}