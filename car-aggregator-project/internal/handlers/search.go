package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"car-aggregator/internal/services"
)

type DadataServiceInterface interface {
	NormalizeCarName(query string) (string, error)
}

type CarAPIServiceInterface interface {
	SearchCars(query string) ([]services.CarData, error)
	GetTrims(carID int) ([]services.Trim, error)
}

type SearchHandler struct {
	dadataService DadataServiceInterface
	carapiService CarAPIServiceInterface
}

type SearchRequest struct {
	Q string `json:"q" binding:"required"`
}

type SearchResponse struct {
	Info string `json:"info"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func NewSearchHandler(dadataService DadataServiceInterface, carapiService CarAPIServiceInterface) *SearchHandler {
	return &SearchHandler{
		dadataService: dadataService,
		carapiService: carapiService,
	}
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.Q) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Empty search query",
			Message: "Query parameter 'q' cannot be empty",
		})
		return
	}

	normalizedQuery, err := h.dadataService.NormalizeCarName(req.Q)
	if err != nil {
		normalizedQuery = req.Q
	}

	cars, err := h.carapiService.SearchCars(normalizedQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to search cars",
			Message: err.Error(),
		})
		return
	}

	info := h.formatSearchResponse(cars)

	c.JSON(http.StatusOK, SearchResponse{
		Info: info,
	})
}

func (h *SearchHandler) SearchTrims(c *gin.Context) {
	carIDStr := c.Query("car_id")
	if carIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing car_id parameter",
			Message: "car_id query parameter is required",
		})
		return
	}

	carID, err := strconv.Atoi(carIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid car_id",
			Message: "car_id must be a valid integer",
		})
		return
	}

	trims, err := h.carapiService.GetTrims(carID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get trims",
			Message: err.Error(),
		})
		return
	}

	info := h.formatTrimsResponse(trims)

	c.JSON(http.StatusOK, SearchResponse{
		Info: info,
	})
}

func (h *SearchHandler) formatSearchResponse(cars []services.CarData) string {
	if len(cars) == 0 {
		return "No vehicles found for the given query."
	}

	var result strings.Builder
	result.WriteString("Vehicle Information:\n\n")

	carMap := make(map[string][]services.CarData)
	for _, car := range cars {
		key := fmt.Sprintf("%s %s", car.Make, car.Model)
		carMap[key] = append(carMap[key], car)
	}

	totalTrims := 0
	var trimDetails strings.Builder

	for _, carList := range carMap {
		for _, car := range carList {
			trims, err := h.carapiService.GetTrims(car.ID)
			if err != nil {
				continue
			}

			totalTrims += len(trims)

			for _, trim := range trims {
				trimDetails.WriteString(fmt.Sprintf("- %d %s %s (MSRP: $%d)\n %s\n",
					car.Year, car.Make, car.Model, trim.MSRP, trim.Description))
			}
		}
	}

	result.WriteString(fmt.Sprintf("Available Trims (Total: %d):\n", totalTrims))
	result.WriteString(trimDetails.String())

	return result.String()
}

func (h *SearchHandler) formatTrimsResponse(trims []services.Trim) string {
	if len(trims) == 0 {
		return "No trims found for the given car."
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Available Trims (Total: %d):\n", len(trims)))

	for _, trim := range trims {
		result.WriteString(fmt.Sprintf("- %s (MSRP: $%d)\n %s\n",
			trim.Name, trim.MSRP, trim.Description))
	}

	return result.String()
}