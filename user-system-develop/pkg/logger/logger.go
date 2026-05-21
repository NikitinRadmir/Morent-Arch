package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New создаёт структурированный slog-логгер для stdout.
// LOG_LEVEL: debug | info | warn | error (по умолчанию info).
// LOG_FORMAT: json | text (по умолчанию json).
func New() *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	format := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_FORMAT")))
	if format == "" {
		format = "json"
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	var h slog.Handler
	switch format {
	case "text":
		h = slog.NewTextHandler(os.Stdout, opts)
	default:
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(h)
}

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
