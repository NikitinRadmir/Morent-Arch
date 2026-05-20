package morentevents

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	SchemaVersion = "1"
	SourceMorent  = "morent-backend"
)

// Envelope — единый формат событий Morent ↔ user-system.
type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	SchemaVersion string          `json:"schema_version"`
	Timestamp     time.Time       `json:"timestamp"`
	Source        string          `json:"source"`
	Data          json.RawMessage `json:"data"`
}

func MarshalEnvelope(eventID, eventType string, data any) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal event data: %w", err)
	}
	env := Envelope{
		EventID:       eventID,
		EventType:     eventType,
		SchemaVersion: SchemaVersion,
		Timestamp:     time.Now().UTC(),
		Source:        SourceMorent,
		Data:          raw,
	}
	out, err := json.Marshal(env)
	if err != nil {
		return nil, fmt.Errorf("marshal envelope: %w", err)
	}
	return out, nil
}

func UnmarshalEnvelope(raw []byte) (*Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("unmarshal envelope: %w", err)
	}
	if env.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported schema version: %s", env.SchemaVersion)
	}
	if env.Source != SourceMorent {
		return nil, fmt.Errorf("unexpected event source: %s", env.Source)
	}
	if env.EventID == "" || env.EventType == "" {
		return nil, fmt.Errorf("invalid envelope: missing event_id or event_type")
	}
	return &env, nil
}

func UnmarshalData[T any](env *Envelope) (*T, error) {
	var data T
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshal event data: %w", err)
	}
	return &data, nil
}
