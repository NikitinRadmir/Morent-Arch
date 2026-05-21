package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	morentevents "morent-events"
	"morent-backend/internal/messaging"
)

type EmailPublisher struct {
	writer *kafka.Writer
	topic  string
	log    *slog.Logger
}

func NewEmailPublisher(brokers, topic string, log *slog.Logger) (*EmailPublisher, error) {
	brokers = strings.TrimSpace(brokers)
	if brokers == "" {
		return nil, fmt.Errorf("kafka brokers are empty")
	}
	if topic == "" {
		topic = morentevents.TopicEmails
	}
	if log == nil {
		log = slog.Default()
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(strings.Split(brokers, ",")...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireAll,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}

	return &EmailPublisher{writer: writer, topic: topic, log: log}, nil
}

func (p *EmailPublisher) Enabled() bool {
	return p != nil && p.writer != nil
}

func (p *EmailPublisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func (p *EmailPublisher) PublishEmailSend(ctx context.Context, in messaging.EmailSendEvent) error {
	if !p.Enabled() {
		return nil
	}
	to := strings.ToLower(strings.TrimSpace(in.To))
	if to == "" {
		return fmt.Errorf("email recipient is empty")
	}
	if strings.TrimSpace(in.TemplateKey) == "" {
		return fmt.Errorf("email template key is empty")
	}

	correlationID := strings.TrimSpace(in.CorrelationID)
	if correlationID == "" {
		correlationID = strings.ReplaceAll(uuid.NewString(), "-", "")
	}

	payload := morentevents.EmailSendRequested{
		CorrelationID: correlationID,
		To:            to,
		TemplateKey:   in.TemplateKey,
		Variables:     in.Variables,
		Priority:      in.Priority,
	}
	if payload.Variables == nil {
		payload.Variables = map[string]string{}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	body, err := morentevents.MarshalEnvelope(uuid.NewString(), morentevents.EventEmailSend, payload)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(to),
		Value: body,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		p.log.Warn("kafka email publish failed", "template", in.TemplateKey, "error", err)
		return err
	}
	p.log.Info("kafka email event published", "template", in.TemplateKey, "topic", p.topic)
	return nil
}
