package health

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Report struct {
	Status     string            `json:"status"`
	Service    string            `json:"service"`
	Components map[string]string `json:"components"`
}

func CheckPostgres(ctx context.Context, db *gorm.DB) Report {
	report := Report{
		Service:    "payment-service",
		Components: map[string]string{},
		Status:     "ok",
	}
	if db == nil {
		report.Status = "unavailable"
		report.Components["postgres"] = "not_configured"
		return report
	}
	sqlDB, err := db.DB()
	if err != nil {
		report.Status = "unavailable"
		report.Components["postgres"] = err.Error()
		return report
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		report.Status = "unavailable"
		report.Components["postgres"] = err.Error()
		return report
	}
	report.Components["postgres"] = "up"
	return report
}

func (r Report) HTTPStatus() int {
	if r.Status == "ok" {
		return 200
	}
	return 503
}

const CheckTimeout = 3 * time.Second
