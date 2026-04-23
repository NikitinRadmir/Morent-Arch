package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"car-aggregator/internal/services"
)

type MockDadataService struct {
	mock.Mock
}

func (m *MockDadataService) NormalizeCarName(query string) (string, error) {
	args := m.Called(query)
	return args.String(0), args.Error(1)
}

type MockCarAPIService struct {
	mock.Mock
}

func (m *MockCarAPIService) SearchCars(query string) ([]services.CarData, error) {
	args := m.Called(query)
	return args.Get(0).([]services.CarData), args.Error(1)
}

func (m *MockCarAPIService) GetTrims(carID int) ([]services.Trim, error) {
	args := m.Called(carID)
	return args.Get(0).([]services.Trim), args.Error(1)
}

func TestSearchHandler_Search(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		setupMocks     func(*MockDadataService, *MockCarAPIService)
	}{
		{
			name: "successful search",
			requestBody: SearchRequest{
				Q: "Volkswagen Golf",
			},
			expectedStatus: http.StatusOK,
			setupMocks: func(dadata *MockDadataService, carapi *MockCarAPIService) {
				dadata.On("NormalizeCarName", "Volkswagen Golf").Return("Volkswagen Golf", nil)
				carapi.On("SearchCars", "Volkswagen Golf").Return([]services.CarData{
					{
						ID:    1,
						Make:  "Volkswagen",
						Model: "Golf",
						Year:  2015,
					},
				}, nil)
				carapi.On("GetTrims", 1).Return([]services.Trim{
					{
						ID:          1,
						Name:        "TSI S",
						Description: "TSI S 2dr Hatchback (1.8L 4cyl Turbo 5M)",
						MSRP:        19295,
					},
				}, nil)
			},
		},
		{
			name: "empty query",
			requestBody: SearchRequest{
				Q: "",
			},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockDadataService, *MockCarAPIService) {},
		},
		{
			name:           "invalid request body",
			requestBody:    map[string]interface{}{"invalid": "data"},
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockDadataService, *MockCarAPIService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDadata := new(MockDadataService)
			mockCarAPI := new(MockCarAPIService)
			tt.setupMocks(mockDadata, mockCarAPI)

			handler := NewSearchHandler(mockDadata, mockCarAPI)

			router := gin.New()
			router.GET("/search", handler.Search)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("GET", "/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response SearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Info)
			}

			mockDadata.AssertExpectations(t)
			mockCarAPI.AssertExpectations(t)
		})
	}
}

func TestSearchHandler_SearchTrims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		carID          string
		expectedStatus int
		setupMocks     func(*MockCarAPIService)
	}{
		{
			name:           "successful trims search",
			carID:          "123",
			expectedStatus: http.StatusOK,
			setupMocks: func(carapi *MockCarAPIService) {
				carapi.On("GetTrims", 123).Return([]services.Trim{
					{
						ID:          1,
						Name:        "TSI S",
						Description: "TSI S 2dr Hatchback (1.8L 4cyl Turbo 5M)",
						MSRP:        19295,
					},
				}, nil)
			},
		},
		{
			name:           "missing car_id",
			carID:          "",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockCarAPIService) {},
		},
		{
			name:           "invalid car_id",
			carID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(*MockCarAPIService) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCarAPI := new(MockCarAPIService)
			tt.setupMocks(mockCarAPI)

			handler := NewSearchHandler(nil, mockCarAPI)

			router := gin.New()
			router.GET("/search/trims", handler.SearchTrims)

			url := "/search/trims"
			if tt.carID != "" {
				url += "?car_id=" + tt.carID
			}
			req := httptest.NewRequest("GET", url, nil)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response SearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Info)
			}

			mockCarAPI.AssertExpectations(t)
		})
	}
}