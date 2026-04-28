package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"car-aggregator/internal/config"
	"car-aggregator/internal/db"
	"car-aggregator/internal/handlers"
	"car-aggregator/internal/logging"
	"car-aggregator/internal/models"
	"car-aggregator/internal/repositories"
	"car-aggregator/internal/services"
	"car-aggregator/internal/validators"
)

func getenv(k string, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("env file not loaded: %v", err)
	}

	// Load search configuration
	searchConfig, err := config.LoadSearchConfig()
	if err != nil {
		log.Fatalf("failed to load search config: %v", err)
	}

	// Initialize logger
	logger := logging.NewSearchLogger(searchConfig.Logging.Level, searchConfig.Logging.EnableDebugInfo)

	dsn := getenv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable")

	gormDB, err := db.New(db.Config{DatabaseURL: dsn})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := gormDB.WithContext(context.Background()).AutoMigrate(&models.Offer{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	offerRepo := repositories.NewOfferRepository(gormDB)

	// Initialize enhanced clients
	dadataClient := services.NewEnhancedDaDataClient(searchConfig, logger)
	carAPIClient := services.NewEnhancedCarAPIClient(
		os.Getenv("CARAPI_TOKEN"),
		os.Getenv("CARAPI_SECRET"),
		searchConfig,
		logger,
	)

	searchValidator := validators.NewUserSearchRequestValidator()

	// Create legacy search service for backward compatibility
	legacySearchService := services.NewSearchService(
		offerRepo,
		searchValidator,
	)

	// Create enhanced search service
	enhancedSearchService := services.NewEnhancedSearchService(
		searchConfig,
		logger,
		dadataClient,
		carAPIClient,
		offerRepo,
		searchValidator,
		legacySearchService,
	)

	searchHandler := handlers.NewSearchHandler(enhancedSearchService, offerRepo)
	offerHandler := handlers.NewOfferHandler(offerRepo)

	router := handlers.NewRouter(searchHandler, offerHandler)
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
