package logger

import (
	"log/slog"
	"os"
)

// New создаёт JSON-логгер для stdout (структурированные логи).
func New() *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	return slog.New(h)
}
