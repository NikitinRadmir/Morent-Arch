package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"car-aggregator/internal/handlers"
	"car-aggregator/internal/services"
)

func getenv(k string, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("env file not loaded: %v", err)
	}

	config := &Config{
		HTTPAddr:        getenv("HTTP_ADDR", ":8080"),
		DadataAPIKey:    getenv("DADATA_API_KEY", ""),
		DadataSecretKey: getenv("DADATA_SECRET_KEY", ""),
		CarAPIToken:     getenv("CARAPI_TOKEN", ""),
		CarAPISecret:    getenv("CARAPI_SECRET", ""),
	}

	dadataService := services.NewDadataService(config.DadataAPIKey, config.DadataSecretKey)
	carapiService := services.NewCarAPIService(config.CarAPIToken, config.CarAPISecret)

	searchHandler := handlers.NewSearchHandler(dadataService, carapiService)

	router := handlers.NewRouter(searchHandler)

	log.Printf("Starting server on %s", config.HTTPAddr)
	if err := router.Run(config.HTTPAddr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

type Config struct {
	HTTPAddr        string
	DadataAPIKey    string
	DadataSecretKey string
	CarAPIToken     string
	CarAPISecret    string
}