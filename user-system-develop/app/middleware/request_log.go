package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLog пишет структурированный лог каждого HTTP-запроса.
func RequestLog(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		attrs := []any{
			"log_type", "http",
			"service", "user-system",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"duration_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if query != "" {
			attrs = append(attrs, "query", query)
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		switch {
		case status >= 500:
			log.Error("http request", attrs...)
		case status >= 400:
			log.Warn("http request", attrs...)
		default:
			log.Info("http request", attrs...)
		}
	}
}

// Recover перехватывает panic и пишет структурированный лог.
func Recover(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recovered",
					"panic", rec,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"client_ip", c.ClientIP(),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
