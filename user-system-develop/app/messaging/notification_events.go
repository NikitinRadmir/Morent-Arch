package messaging

import (
	"encoding/json"
	"time"
)

type NotificationEvent struct {
	CompanyID string `json:"company_id"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

type RabbitMQEnvelope struct {
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}
