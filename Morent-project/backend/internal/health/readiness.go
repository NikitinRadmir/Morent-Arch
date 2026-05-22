package health

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ComponentStatus — состояние зависимости.
type ComponentStatus struct {
	Status  string `json:"status"` // up | down
	Message string `json:"message,omitempty"`
}

// Report — результат /ready.
type Report struct {
	Status     string                     `json:"status"` // ok | degraded | unavailable
	Service    string                     `json:"service"`
	Components map[string]ComponentStatus `json:"components"`
}

// Checker проверяет критичные зависимости Morent backend.
type Checker struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (c *Checker) Check(ctx context.Context) Report {
	report := Report{
		Service:    "morent-backend",
		Components: make(map[string]ComponentStatus),
		Status:     "ok",
	}
	if c == nil {
		report.Status = "unavailable"
		return report
	}

	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			report.Components["postgres"] = ComponentStatus{Status: "down", Message: err.Error()}
		} else if err := sqlDB.PingContext(ctx); err != nil {
			report.Components["postgres"] = ComponentStatus{Status: "down", Message: err.Error()}
		} else {
			report.Components["postgres"] = ComponentStatus{Status: "up"}
		}
	} else {
		report.Components["postgres"] = ComponentStatus{Status: "down", Message: "not configured"}
	}

	if c.Redis != nil {
		if err := c.Redis.Ping(ctx).Err(); err != nil {
			report.Components["redis"] = ComponentStatus{Status: "down", Message: err.Error()}
		} else {
			report.Components["redis"] = ComponentStatus{Status: "up"}
		}
	} else {
		report.Components["redis"] = ComponentStatus{Status: "down", Message: "not configured"}
	}

	for _, st := range report.Components {
		if st.Status != "up" {
			report.Status = "unavailable"
			break
		}
	}
	return report
}

// HTTPStatus возвращает 200 или 503 для readiness probe.
func (r Report) HTTPStatus() int {
	if r.Status == "ok" {
		return 200
	}
	return 503
}

// DefaultCheckTimeout для Ping зависимостей.
const DefaultCheckTimeout = 3 * time.Second
