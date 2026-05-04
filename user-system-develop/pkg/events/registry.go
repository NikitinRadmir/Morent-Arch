package events

import (
	"context"
	"encoding/json"
	"fmt"
)

// HandlerFunc processes a raw Envelope.
type HandlerFunc func(ctx context.Context, env Envelope) error

// Registry maps event types to typed handlers.
type Registry struct {
	handlers map[string]HandlerFunc
}

// NewRegistry creates an empty event registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]HandlerFunc)}
}

// Register adds a typed handler for the given event type.
// When Handle is called, the registry automatically unmarshals
// Envelope.Data into T and calls the handler.
func Register[T any](r *Registry, eventType string, handler func(ctx context.Context, event T) error) {
	r.handlers[eventType] = func(ctx context.Context, env Envelope) error {
		var data T
		if err := json.Unmarshal(env.Data, &data); err != nil {
			return fmt.Errorf("unmarshal %q event data: %w", eventType, err)
		}
		return handler(ctx, data)
	}
}

// Handle parses the raw message, looks up the event type, and dispatches
// to the appropriate typed handler.
func (r *Registry) Handle(ctx context.Context, raw []byte) error {
	env, err := ParseEnvelope(raw)
	if err != nil {
		return fmt.Errorf("parse envelope: %w", err)
	}

	handler, ok := r.handlers[env.EventType]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownEventType, env.EventType)
	}
	return handler(ctx, *env)
}
