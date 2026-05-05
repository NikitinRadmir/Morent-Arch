package events

// Event type constants for inter-service communication.
const (
	EventUserDeleted      = "user.deleted"
	EventNotificationSend = "notification.send"
)

// UserDeletedEvent is published by user-system when a user
// is deleted.
type UserDeletedEvent struct {
	UserID string `json:"user_id"`
}

// NotificationEvent is consumed by notification-service
// to send an email.
type NotificationEvent struct {
	Transport string `json:"transport"`
	To        string `json:"to"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}
