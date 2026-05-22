package morentevents

const (
	TopicEmails = "morent.emails"

	EventEmailSend = "morent.email.send"

	EmailTemplateBookingConfirmation = "booking_confirmation"
	EmailTemplateWelcomeRegistered   = "welcome_registered"
	EmailTemplateEmailVerification   = "email_verification"
	EmailTemplateLoginNotification   = "login_notification"
	EmailTemplatePaymentFailed       = "payment_failed"
	EmailTemplateReminder24h         = "reminder_24h"
)

// EmailSendRequested — запрос на отправку письма (Morent → EmailService).
type EmailSendRequested struct {
	CorrelationID string            `json:"correlationId"`
	To            string            `json:"to"`
	TemplateKey   string            `json:"templateKey"`
	Variables     map[string]string `json:"variables"`
	Priority      string            `json:"priority,omitempty"`
}
