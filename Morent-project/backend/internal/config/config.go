package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	MinioEndpoint       string `json:"minio_endpoint"`
	MinioAccessKey      string `json:"minio_access_key"`
	MinioSecretKey      string `json:"minio_secret_key"`
	MinioBucket         string `json:"minio_bucket"`
	MinioPublicEndpoint string `json:"minio_public_endpoint"`
	MinioUseSSL         bool   `json:"minio_use_ssl"`

	GRPCPort string `json:"grpc_port"`

	FrontendOrigin       string `json:"frontend_origin"`
	AggregatorTimeoutSec int    `json:"aggregator_timeout_sec"`
	AggregatorRetryCount int    `json:"aggregator_retry_count"`
	SessionCookieName    string `json:"session_cookie_name"`
	SessionCookieSecure  bool   `json:"session_cookie_secure"`
	SessionCookieDomain  string `json:"session_cookie_domain"`
}

func LoadConfig(_ string) (*Config, error) {
	cfg := &Config{
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5433"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgresmaster"),
		DBName:               getEnv("DB_NAME", "car_booking"),
		MinioEndpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey:       getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:       getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:          getEnv("MINIO_BUCKET", "morent-media"),
		MinioPublicEndpoint:  getEnv("MINIO_PUBLIC_ENDPOINT", ""),
		GRPCPort:             getEnv("GRPC_PORT", "50051"),
		FrontendOrigin:       getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
		AggregatorTimeoutSec: getEnvInt("AGGREGATOR_TIMEOUT_SEC", 35),
		AggregatorRetryCount: getEnvInt("AGGREGATOR_RETRY_COUNT", 3),
		SessionCookieName:    getEnv("SESSION_COOKIE_NAME", "morent_session"),
		SessionCookieDomain:  getEnv("SESSION_COOKIE_DOMAIN", ""),
	}
	cfg.MinioUseSSL = getEnvBool("MINIO_USE_SSL", false)
	cfg.SessionCookieSecure = getEnvBool("SESSION_COOKIE_SECURE", false)

	if cfg.MinioPublicEndpoint == "" {
		cfg.MinioPublicEndpoint = cfg.MinioEndpoint
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, parseBoolErr := strconv.ParseBool(value)
	if parseBoolErr != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, parseIntErr := strconv.Atoi(value)
	if parseIntErr != nil {
		return fallback
	}
	return parsed
}
