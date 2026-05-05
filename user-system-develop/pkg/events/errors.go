package events

import "errors"

// ErrUnknownEventType is returned when no handler is registered
// for the given event type.
var ErrUnknownEventType = errors.New("unknown event type")
