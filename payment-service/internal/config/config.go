package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv  string
	HTTPAddr string

	KafkaEnabled           bool
	KafkaBrokers           string
	KafkaTopicBankCommands string
	KafkaTopicBankResponses string
	KafkaGroupPaymentBank  string
}

func Load() Config {
	loadDotEnv()
	brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	return Config{
		AppEnv:                  getEnv("APP_ENV", "dev"),
		HTTPAddr:                getEnv("HTTP_ADDR", ":8081"),
		KafkaBrokers:            brokers,
		KafkaTopicBankCommands:  getEnv("KAFKA_TOPIC_BANK_COMMANDS", "morent.bank.commands"),
		KafkaTopicBankResponses: getEnv("KAFKA_TOPIC_BANK_RESPONSES", "morent.bank.responses"),
		KafkaGroupPaymentBank:   getEnv("KAFKA_GROUP_PAYMENT_BANK", "payment-service-bank"),
		KafkaEnabled:            getEnvBool("KAFKA_ENABLED", brokers != ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
