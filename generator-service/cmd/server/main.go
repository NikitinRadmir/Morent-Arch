package main

import (
	"generator-service/internal/handlers"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Настройка Gin режима
	gin.SetMode(gin.ReleaseMode)
	
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		if path == "/health" {
			return
		}
		status := c.Writer.Status()
		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(c.Request.Context(), level, "http_request",
			"log_type", "http",
			"service", "generator-service",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})

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
		api.POST("/password/validate", handler.ValidatePassword)
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
