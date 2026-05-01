package domain

import "time"

// OperationLimits хранит информацию о лимитах операций за период.
type OperationLimits struct {
	AccountID   string
	Period      LimitPeriod
	SpentAmount Money
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type LimitPeriod string

const (
	LimitDaily   LimitPeriod = "daily"
	LimitMonthly LimitPeriod = "monthly"
)

// CheckLimit проверяет, не превышен ли лимит.
func (l *OperationLimits) CheckLimit(limit Money, amount Money) bool {
	return l.SpentAmount+amount <= limit
}
