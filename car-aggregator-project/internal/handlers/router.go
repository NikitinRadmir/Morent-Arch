package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(searchHandler *SearchHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "car-aggregator",
		})
	})

	router.GET("/search", searchHandler.Search)
	router.GET("/search/trims", searchHandler.SearchTrims)

	return router
}