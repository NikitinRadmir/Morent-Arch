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
	"time"

	"car-aggregator/internal/dtos"
)

type CarAPIClient interface {
	Login(ctx context.Context) error
	GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error)
	GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error)
}

type carAPIClient struct {
	baseURL    string
	apiToken   string
	apiSecret  string
	jwtToken   string
	httpClient *http.Client
}

func NewCarAPIClient(apiToken, apiSecret string) CarAPIClient {
	return &carAPIClient{
		baseURL:   "https://carapi.app/api",
		apiToken:  apiToken,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *carAPIClient) Login(ctx context.Context) error {
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

func (c *carAPIClient) GetTrimsByModel(ctx context.Context, model string, limit int) (*dtos.TrimsResponse, error) {
	if c.jwtToken == "" {
		if err := c.Login(ctx); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	url := fmt.Sprintf("%s/trims/v2?limit=%d&model=%s", c.baseURL, limit, url.QueryEscape(model))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.jwtToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	var trimsResp dtos.TrimsResponse
	if err := json.NewDecoder(resp.Body).Decode(&trimsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &trimsResp, nil
}

func (c *carAPIClient) GetVehicleSpecs(ctx context.Context, year int, make, model string) (int, float64, error) {
	if c.jwtToken == "" {
		if err := c.Login(ctx); err != nil {
			return 0, 0, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	seats, err := c.getSeats(ctx, year, make, model)
	if err != nil {
		return 0, 0, err
	}

	fuel, err := c.getFuelConsumption(ctx, year, make, model)
	if err != nil {
		return 0, 0, err
	}

	return seats, fuel, nil
}

func (c *carAPIClient) getSeats(ctx context.Context, year int, make, model string) (int, error) {
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

func (c *carAPIClient) getFuelConsumption(ctx context.Context, year int, make, model string) (float64, error) {
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

	// Приближенная оценка, когда MPG в API недоступен.
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
