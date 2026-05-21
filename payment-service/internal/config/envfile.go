package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// loadDotEnv подхватывает .env из текущей директории или корня payment-service.
func loadDotEnv() {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	candidates := []string{
		filepath.Join(wd, ".env"),
		filepath.Join(wd, "..", ".env"),
		".env",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		// Overload: значения из .env имеют приоритет над пустыми/старыми переменными shell.
		if err := godotenv.Overload(path); err != nil {
			slog.Warn("payment-service: .env parse failed", "path", path, "error", err)
			continue
		}
		slog.Info("payment-service: loaded .env", "path", path)
		return
	}

	slog.Warn("payment-service: .env not found", "cwd", wd)
}
