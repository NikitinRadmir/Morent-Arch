package logging

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

type SearchLogger interface {
	LogSearchStart(query string)
	LogDaDataResult(query string, results []DaDataResult)
	LogCarAPIRequest(make, model string, limit int)
	LogCarAPIResponse(count int, duration time.Duration)
	LogFallbackUsed(strategy string, reason string)
	LogSearchComplete(query string, resultCount int, duration time.Duration)
	LogError(operation string, err error, context map[string]interface{})
}

type DaDataResult struct {
	Brand      string  `json:"brand"`
	Model      string  `json:"model"`
	Confidence float64 `json:"confidence"`
}

type CarAPIRequest struct {
	Make     string    `json:"make"`
	Model    string    `json:"model"`
	Limit    int       `json:"limit"`
	Timestamp time.Time `json:"timestamp"`
}

type searchLogger struct {
	level       string
	enableDebug bool
}

func NewSearchLogger(level string, enableDebug bool) SearchLogger {
	return &searchLogger{
		level:       level,
		enableDebug: enableDebug,
	}
}

func (l *searchLogger) LogSearchStart(query string) {
	l.logInfo("search_start", map[string]interface{}{
		"query":     query,
		"timestamp": time.Now(),
	})
}

func (l *searchLogger) LogDaDataResult(query string, results []DaDataResult) {
	l.logInfo("dadata_result", map[string]interface{}{
		"query":        query,
		"results_count": len(results),
		"results":      results,
		"timestamp":    time.Now(),
	})
}

func (l *searchLogger) LogCarAPIRequest(make, model string, limit int) {
	l.logInfo("carapi_request", map[string]interface{}{
		"make":      make,
		"model":     model,
		"limit":     limit,
		"timestamp": time.Now(),
	})
}

func (l *searchLogger) LogCarAPIResponse(count int, duration time.Duration) {
	l.logInfo("carapi_response", map[string]interface{}{
		"results_count": count,
		"duration_ms":   duration.Milliseconds(),
		"timestamp":     time.Now(),
	})
}

func (l *searchLogger) LogFallbackUsed(strategy string, reason string) {
	l.logInfo("fallback_used", map[string]interface{}{
		"strategy":  strategy,
		"reason":    reason,
		"timestamp": time.Now(),
	})
}

func (l *searchLogger) LogSearchComplete(query string, resultCount int, duration time.Duration) {
	l.logInfo("search_complete", map[string]interface{}{
		"query":        query,
		"result_count": resultCount,
		"duration_ms":  duration.Milliseconds(),
		"timestamp":    time.Now(),
	})
}

func (l *searchLogger) LogError(operation string, err error, context map[string]interface{}) {
	logData := map[string]interface{}{
		"operation": operation,
		"error":     err.Error(),
		"timestamp": time.Now(),
	}
	
	for k, v := range context {
		logData[k] = v
	}
	
	l.logError("search_error", logData)
}

func (l *searchLogger) logInfo(event string, data map[string]interface{}) {
	if l.level == "debug" || l.level == "info" {
		l.logEvent("INFO", event, data)
	}
}

func (l *searchLogger) logError(event string, data map[string]interface{}) {
	l.logEvent("ERROR", event, data)
}

func (l *searchLogger) logEvent(level, event string, data map[string]interface{}) {
	logEntry := map[string]interface{}{
		"level": level,
		"event": event,
		"data":  data,
	}
	
	if jsonData, err := json.Marshal(logEntry); err == nil {
		log.Printf("[SEARCH_LOG] %s", string(jsonData))
	} else {
		log.Printf("[SEARCH_LOG] Failed to marshal log entry: %v", err)
	}
}

// Context-aware logger
type ContextLogger struct {
	logger SearchLogger
	ctx    context.Context
}

func WithContext(ctx context.Context, logger SearchLogger) *ContextLogger {
	return &ContextLogger{
		logger: logger,
		ctx:    ctx,
	}
}

func (cl *ContextLogger) LogSearchStart(query string) {
	cl.logger.LogSearchStart(query)
}

func (cl *ContextLogger) LogDaDataResult(query string, results []DaDataResult) {
	cl.logger.LogDaDataResult(query, results)
}

func (cl *ContextLogger) LogCarAPIRequest(make, model string, limit int) {
	cl.logger.LogCarAPIRequest(make, model, limit)
}

func (cl *ContextLogger) LogCarAPIResponse(count int, duration time.Duration) {
	cl.logger.LogCarAPIResponse(count, duration)
}

func (cl *ContextLogger) LogFallbackUsed(strategy string, reason string) {
	cl.logger.LogFallbackUsed(strategy, reason)
}

func (cl *ContextLogger) LogSearchComplete(query string, resultCount int, duration time.Duration) {
	cl.logger.LogSearchComplete(query, resultCount, duration)
}

func (cl *ContextLogger) LogError(operation string, err error, context map[string]interface{}) {
	cl.logger.LogError(operation, err, context)
}