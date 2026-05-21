package config

import (
	"os"
	"strconv"
	"strings"
)

// Config загружается только из переменных окружения (см. .env.example в корне Morent-project).
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	MinioEndpoint       string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioBucket         string
	MinioPublicEndpoint string
	MinioUseSSL         bool

	GRPCPort string

	FrontendOrigin        string
	AggregatorBaseURL     string
	AggregatorTimeoutSec  int
	AggregatorRetryCount  int
	GeneratorBaseURL      string
	SessionCookieName     string
	SessionCookieSecure   bool
	SessionCookieDomain   string
	BankSessionCookieName string
	BankLinkSecret        string

	RedisAddr             string
	RedisPassword         string
	RedisDB               int
	CarsCacheTTLSeconds   int

	KafkaEnabled              bool
	KafkaBrokers              string
	KafkaTopicUsers           string
	KafkaTopicEmails          string
	KafkaTopicBankCommands    string
	KafkaTopicBankResponses   string
	KafkaGroupMorentBank      string
	MorentCompanyName         string
}

// LoadFromEnv читает конфигурацию из окружения. Файл config.json не используется.
func LoadFromEnv() (*Config, error) {
	loadDotEnv()
	baseURL := strings.TrimRight(getEnv("AGGREGATOR_BASE_URL", "http://localhost:8080"), "/")

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
		AggregatorBaseURL:    baseURL,
		AggregatorTimeoutSec: getEnvInt("AGGREGATOR_TIMEOUT_SEC", 35),
		AggregatorRetryCount: getEnvInt("AGGREGATOR_RETRY_COUNT", 3),
		GeneratorBaseURL:     strings.TrimRight(getEnv("GENERATOR_BASE_URL", "http://localhost:8080"), "/"),
		SessionCookieName:     getEnv("SESSION_COOKIE_NAME", "morent_session"),
		SessionCookieDomain:   getEnv("SESSION_COOKIE_DOMAIN", ""),
		BankSessionCookieName: getEnv("BANK_SESSION_COOKIE_NAME", "morent_bank_session"),
		BankLinkSecret:        getEnv("BANK_LINK_SECRET", "morent-bank-link-dev"),
		RedisAddr:            getEnv("REDIS_ADDR", ""),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              getEnvInt("REDIS_DB", 0),
		CarsCacheTTLSeconds:  getEnvInt("CARS_CACHE_TTL_SEC", 60),
		KafkaBrokers:            getEnv("KAFKA_BROKERS", ""),
		KafkaTopicUsers:         getEnv("KAFKA_TOPIC_USERS", "morent.users"),
		KafkaTopicEmails:        getEnv("KAFKA_TOPIC_EMAILS", "morent.emails"),
		KafkaTopicBankCommands:  getEnv("KAFKA_TOPIC_BANK_COMMANDS", "morent.bank.commands"),
		KafkaTopicBankResponses: getEnv("KAFKA_TOPIC_BANK_RESPONSES", "morent.bank.responses"),
		KafkaGroupMorentBank:    getEnv("KAFKA_GROUP_MORENT_BANK", "morent-backend-bank"),
		MorentCompanyName:       getEnv("MORENT_COMPANY_NAME", "Morent"),
	}
	cfg.KafkaEnabled = getEnvBool("KAFKA_ENABLED", cfg.KafkaBrokers != "")
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
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
