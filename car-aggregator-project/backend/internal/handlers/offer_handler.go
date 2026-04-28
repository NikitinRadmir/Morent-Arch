package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"car-aggregator/internal/errors"
	"car-aggregator/internal/models"
	"car-aggregator/internal/repositories"
)

type OfferHandler struct {
	offerRepo repositories.OfferRepository
}

func NewOfferHandler(offerRepo repositories.OfferRepository) *OfferHandler {
	return &OfferHandler{
		offerRepo: offerRepo,
	}
}

// PurchaseOffer handles POST /offers/purchase
func (h *OfferHandler) PurchaseOffer(c *gin.Context) {
	var req models.PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request: "+err.Error()))
		return
	}

	err := h.offerRepo.PurchaseOffer(c.Request.Context(), req)
	if err != nil {
		RenderError(c, errors.New(errors.CodeInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "Offer created successfully",
		"car_id":   req.CarID,
		"service":  req.CustomerService,
	})
}

// ReleaseOffer handles POST /offers/release
func (h *OfferHandler) ReleaseOffer(c *gin.Context) {
	var req models.ReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid request: "+err.Error()))
		return
	}

	err := h.offerRepo.ReleaseOffer(c.Request.Context(), req)
	if err != nil {
		RenderError(c, errors.New(errors.CodeInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Offer released successfully",
		"car_id":  req.CarID,
		"service": req.CustomerService,
	})
}

// GetAvailableOffers handles GET /offers/available
func (h *OfferHandler) GetAvailableOffers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"offers": []models.Offer{}, "count": 0})
}

// GetPurchasedOffers handles GET /offers/purchased/:service
func (h *OfferHandler) GetPurchasedOffers(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		RenderError(c, errors.New(errors.CodeValidation, "service name is required"))
		return
	}

	offers, err := h.offerRepo.GetPurchasedOffers(c.Request.Context(), serviceName)
	if err != nil {
		RenderError(c, errors.New(errors.CodeInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"offers":  offers,
		"count":   len(offers),
	})
}

// GetOfferStatus handles GET /offers/:id/status
func (h *OfferHandler) GetOfferStatus(c *gin.Context) {
	idStr := c.Param("id")
	_, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RenderError(c, errors.New(errors.CodeValidation, "invalid offer ID"))
		return
	}
	RenderError(c, errors.New(errors.CodeNotFound, "not implemented"))
}