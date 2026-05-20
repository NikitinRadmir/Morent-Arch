package models

import "time"

// ProcessedEvent — идемпотентная обработка Kafka-событий Morent.
type ProcessedEvent struct {
	EventID     string    `gorm:"primaryKey;size:64"`
	EventType   string    `gorm:"size:64;not null;index"`
	ProcessedAt time.Time `gorm:"not null"`
}
