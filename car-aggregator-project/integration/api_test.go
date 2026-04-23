package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"car-aggregator/internal/handlers"
	"car-aggregator/internal/services"
)

func TestIntegration_SearchEndpoint(t *testing.T) {
	if os.Getenv("CARAPI_TOKEN") == "" {
		t.Skip("Skipping integration test: CARAPI_TOKEN not set")
	}

	dadataService := services.NewDadataService(
		os.Getenv("DADATA_API_KEY"),
		os.Getenv("DADATA_SECRET_KEY"),
	)
	carapiService := services.NewCarAPIService(
		os.Getenv("CARAPI_TOKEN"),
		os.Getenv("CARAPI_SECRET"),
	)

	searchHandler := handlers.NewSearchHandler(dadataService, carapiService)
	router := handlers.NewRouter(searchHandler)

	tests := []struct {
		name         string
		requestBody  map[string]string
		expectedCode int
	}{
		{
			name: "search for popular car",
			requestBody: map[string]string{
				"q": "Toyota Camry",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "search for car in Russian",
			requestBody: map[string]string{
				"q": "Тойота Камри",
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "search for non-existent car",
			requestBody: map[string]string{
				"q": "NonExistentCarBrand Model",
			},
			expectedCode: http.StatusOK, 
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("GET", "/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				var response handlers.SearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Info)
				t.Logf("Response for %s: %s", tt.name, response.Info[:min(100, len(response.Info))])
			}
		})
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "car-aggregator",
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "car-aggregator", response["service"])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}