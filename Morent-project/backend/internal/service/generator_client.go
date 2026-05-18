package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GeneratorClient вызывает generator-service по внутренней сети (пароль только в POST body).
type GeneratorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewGeneratorClient(baseURL string) *GeneratorClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &GeneratorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (c *GeneratorClient) Enabled() bool {
	return c != nil && c.baseURL != ""
}

type generatedPasswordResponse struct {
	Password string `json:"password"`
	Mask     string `json:"mask"`
	Length   int    `json:"length"`
}

// GeneratePassword запрашивает новый пароль у generator-service.
func (c *GeneratorClient) GeneratePassword(ctx context.Context) (password string, err error) {
	if !c.Enabled() {
		return "", fmt.Errorf("generator service is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/password", nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("generator returned status %d", resp.StatusCode)
	}
	var out generatedPasswordResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Password == "" {
		return "", fmt.Errorf("empty password from generator")
	}
	return out.Password, nil
}

// PasswordValidationResult совпадает с ответом generator-service.
type PasswordValidationResult struct {
	Valid      bool     `json:"valid"`
	Errors     []string `json:"errors,omitempty"`
	HasLower   bool     `json:"hasLower"`
	HasUpper   bool     `json:"hasUpper"`
	HasDigit   bool     `json:"hasDigit"`
	HasSpecial bool     `json:"hasSpecial"`
	LengthOK   bool     `json:"lengthOk"`
	Length     int      `json:"length"`
}

// ValidatePassword отправляет пароль на проверку (только POST, без query string).
func (c *GeneratorClient) ValidatePassword(ctx context.Context, password string) (*PasswordValidationResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("generator service is not configured")
	}
	payload, err := json.Marshal(map[string]string{"password": password})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/password/validate", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generator returned status %d", resp.StatusCode)
	}
	var result PasswordValidationResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
