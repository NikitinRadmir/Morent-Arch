package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func loadDotEnv() {
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	candidates := []string{
		filepath.Join(wd, ".env"),
		filepath.Join(wd, "..", ".env"),
		filepath.Join(wd, "../..", ".env"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := godotenv.Overload(path); err != nil {
			slog.Warn("morent-backend: .env parse failed", "path", path, "error", err)
			continue
		}
		slog.Info("morent-backend: loaded .env", "path", path)
		return
	}
}
