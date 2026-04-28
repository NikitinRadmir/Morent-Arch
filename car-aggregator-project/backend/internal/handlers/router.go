package handlers

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(
	searchHandler *SearchHandler,
	offerHandler *OfferHandler,
) *gin.Engine {
	router := gin.Default()

	// Search endpoints
	search := router.Group("/search")
	{
		search.GET("", searchHandler.SearchVehicle)
		search.POST("/trims", searchHandler.SearchTrims)
	}

	// Offer management endpoints
	offers := router.Group("/offers")
	{
		offers.POST("/purchase", offerHandler.PurchaseOffer)
		offers.POST("/release", offerHandler.ReleaseOffer)
		offers.GET("/available", offerHandler.GetAvailableOffers)
		offers.GET("/purchased/:service", offerHandler.GetPurchasedOffers)
		offers.GET("/:id/status", offerHandler.GetOfferStatus)
	}

	return router
}
