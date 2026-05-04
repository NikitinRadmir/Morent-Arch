package events

import (
	"encoding/json"
	"fmt"
	"time"
)

// Envelope is a universal wrapper for events.
type Envelope struct {
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Source    string          `json:"source"`
	Data      json.RawMessage `json:"data"`
}

// Marshal packs a typed event into an Envelope JSON payload.
func Marshal[T any](eventType, source string, data T) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal event data: %w", err)
	}

	env := Envelope{
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Source:    source,
		Data:      raw,
	}

	result, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshal envelope: %w", err)
	}
	return result, nil
}

// Unmarshal extracts a typed event from an Envelope JSON payload.
// Returns an error if the event type does not match expectedType.
func Unmarshal[T any](raw []byte, expectedType string) (*T, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}

	if env.EventType != expectedType {
		return nil, fmt.Errorf(
			"unexpected event type: got %q, want %q", env.EventType, expectedType)
	}

	var data T
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal event data: %w", err)
	}
	return &data, nil
}

// ParseEnvelope extracts just the envelope without unmarshaling the data payload.
func ParseEnvelope(raw []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}
	return &env, nil
}
