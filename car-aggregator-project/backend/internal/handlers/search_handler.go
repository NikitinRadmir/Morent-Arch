package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"car-aggregator/internal/dtos"
	"car-aggregator/internal/errors"
	"car-aggregator/internal/repositories"
	"car-aggregator/internal/services"
)

type SearchHandler struct {
	svc       services.EnhancedSearchService
	offerRepo repositories.OfferRepository
}

func NewSearchHandler(s services.EnhancedSearchService, offerRepo repositories.OfferRepository) *SearchHandler {
	return &SearchHandler{svc: s, offerRepo: offerRepo}
}

func (h *SearchHandler) SearchVehicle(c *gin.Context) {
	var in dtos.UserSearchQuery

	if err := c.ShouldBindJSON(&in); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}

	// Use enhanced search with fallback
	resp, err := h.svc.SearchWithFallback(c.Request.Context(), in)
	if err != nil {
		RenderError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *SearchHandler) SearchTrims(c *gin.Context) {
	var in dtos.UserSearchQuery

	if err := c.ShouldBindJSON(&in); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request"))
		return
	}

	// Use enhanced search service
	resp, err := h.svc.SearchTrims(c.Request.Context(), in)
	if err != nil {
		RenderError(c, err)
		return
	}

	// Отсекаем уже "забранные" машины по таблице offers (car_id).
	ids := make([]int, 0, len(resp.Cars))
	for _, car := range resp.Cars {
		ids = append(ids, car.ID)
	}
	taken, err := h.offerRepo.GetTakenCarIDs(c.Request.Context(), ids)
	if err != nil {
		RenderError(c, errors.New(errors.CodeInternal, err.Error()))
		return
	}
	if len(taken) > 0 {
		filtered := make([]dtos.SearchTrimItem, 0, len(resp.Cars))
		for _, car := range resp.Cars {
			if taken[car.ID] {
				continue
			}
			filtered = append(filtered, car)
		}
		resp.Cars = filtered
		resp.Count = len(filtered)
	}
	c.JSON(http.StatusOK, resp)
}
