package config

import "os"

type Config struct {
	Env      string // dev|prod
	HTTPAddr string // ":8080"
}

func FromEnv() Config {
	return Config{
		Env:      getenv("APP_ENV", "dev"),
		HTTPAddr: getenv("HTTP_ADDR", ":8080"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
