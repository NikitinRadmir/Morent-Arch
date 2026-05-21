package messaging

import "context"

// EmailEventPublisher публикует запросы на отправку email в Kafka (Morent → EmailService).
type EmailEventPublisher interface {
	PublishEmailSend(ctx context.Context, in EmailSendEvent) error
	Enabled() bool
}

type EmailSendEvent struct {
	CorrelationID string
	To            string
	TemplateKey   string
	Variables     map[string]string
	Priority      string
}

// EmailNoopPublisher — заглушка, если Kafka отключён.
type EmailNoopPublisher struct{}

func (EmailNoopPublisher) Enabled() bool { return false }

func (EmailNoopPublisher) PublishEmailSend(context.Context, EmailSendEvent) error {
	return nil
}
