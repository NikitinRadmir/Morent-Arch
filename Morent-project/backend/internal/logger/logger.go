package logger

import (
	"log/slog"

	"obslog"
)

// New создаёт JSON-логгер для stdout (поля service, log_type для Elasticsearch).
func New() *slog.Logger {
	return obslog.New("morent-backend", obslog.LogTypeApp)
}
