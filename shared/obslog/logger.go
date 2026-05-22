package obslog

import (
	"log/slog"
	"os"
)

// LogType — категория записи для индексации в Elasticsearch (поле log.type).
const (
	LogTypeApp        = "app"
	LogTypeHTTP       = "http"
	LogTypeHeartbeat  = "heartbeat"
	LogTypeIntegration = "integration"
	LogTypeHealth     = "health"
)

// New создаёт JSON-логгер с полями service и log.type по умолчанию.
func New(service, defaultLogType string) *slog.Logger {
	if defaultLogType == "" {
		defaultLogType = LogTypeApp
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})
	return slog.New(h).With(
		"service", service,
		"log_type", defaultLogType,
	)
}

// Heartbeat пишет периодический сигнал живости сервиса.
func Heartbeat(log *slog.Logger, component, status string, attrs ...any) {
	if log == nil {
		return
	}
	args := []any{
		"log_type", LogTypeHeartbeat,
		"component", component,
		"status", status,
	}
	args = append(args, attrs...)
	log.Info("heartbeat", args...)
}

// Integration логирует вызов внешнего сервиса / Kafka.
func Integration(log *slog.Logger, target, operation, result string, attrs ...any) {
	if log == nil {
		return
	}
	args := []any{
		"log_type", LogTypeIntegration,
		"target", target,
		"operation", operation,
		"result", result,
	}
	args = append(args, attrs...)
	log.Info("integration", args...)
}
