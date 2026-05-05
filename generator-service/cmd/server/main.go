package main

import (
	"fmt"
	"generator-service/internal/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Настройка Gin режима
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	// Логирование запросов
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s %s %d %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
		)
	}))

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	handler := handlers.NewGeneratorHandler()

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/password", handler.GeneratePassword)
		api.GET("/password/mask", handler.GeneratePasswordByMask)
		api.GET("/qrcode", handler.GenerateQRCode)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "generator-service"})
	})

	log.Println("Generator service starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
