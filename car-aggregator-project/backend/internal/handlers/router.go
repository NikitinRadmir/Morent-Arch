package handlers

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	searchHandler *SearchHandler,
	offerHandler *OfferHandler,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		status := c.Writer.Status()
		if path == "/health" {
			return
		}
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(c.Request.Context(), level, "http_request",
			"log_type", "http",
			"service", "car-aggregator",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "car-aggregator",
		})
	})

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
