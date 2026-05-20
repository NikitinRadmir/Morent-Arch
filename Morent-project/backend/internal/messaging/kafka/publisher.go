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

type Publisher struct {
	writer      *kafka.Writer
	topic       string
	companyName string
	log         *slog.Logger
}

func NewPublisher(brokers, topic, companyName string, log *slog.Logger) (*Publisher, error) {
	brokers = strings.TrimSpace(brokers)
	if brokers == "" {
		return nil, fmt.Errorf("kafka brokers are empty")
	}
	if topic == "" {
		topic = morentevents.TopicUsers
	}
	if companyName == "" {
		companyName = "Morent"
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

	return &Publisher{
		writer:      writer,
		topic:       topic,
		companyName: companyName,
		log:         log,
	}, nil
}

func (p *Publisher) Enabled() bool {
	return p != nil && p.writer != nil
}

func (p *Publisher) Close() error {
	if p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

func (p *Publisher) PublishUserRegistered(ctx context.Context, in messaging.UserRegisteredEvent) error {
	payload := morentevents.UserRegistered{
		MorentUserID: in.MorentUserID,
		Email:        in.Email,
		Name:         in.Name,
		Nickname:     in.Nickname,
		Position:     in.Position,
		AvatarURL:    in.AvatarURL,
		Role:         in.Role,
		PasswordHash: in.PasswordHash,
		CompanyName:  firstNonEmpty(in.CompanyName, p.companyName),
	}
	return p.publish(ctx, morentevents.EventUserRegistered, in.Email, payload)
}

func (p *Publisher) PublishUserProfileUpdated(ctx context.Context, in messaging.UserProfileUpdatedEvent) error {
	payload := morentevents.UserProfileUpdated{
		MorentUserID: in.MorentUserID,
		Email:        in.Email,
		Name:         in.Name,
		Nickname:     in.Nickname,
		Position:     in.Position,
		AvatarURL:    in.AvatarURL,
		Role:         in.Role,
		IsActive:     in.IsActive,
	}
	return p.publish(ctx, morentevents.EventUserProfileUpdated, in.Email, payload)
}

func (p *Publisher) PublishUserPasswordChanged(ctx context.Context, in messaging.UserPasswordChangedEvent) error {
	payload := morentevents.UserPasswordChanged{
		MorentUserID: in.MorentUserID,
		Email:        in.Email,
		PasswordHash: in.PasswordHash,
	}
	return p.publish(ctx, morentevents.EventUserPasswordChanged, in.Email, payload)
}

func (p *Publisher) PublishUserDeactivated(ctx context.Context, in messaging.UserDeactivatedEvent) error {
	payload := morentevents.UserDeactivated{
		MorentUserID: in.MorentUserID,
		Email:        in.Email,
	}
	return p.publish(ctx, morentevents.EventUserDeactivated, in.Email, payload)
}

func (p *Publisher) publish(ctx context.Context, eventType, key string, data any) error {
	if !p.Enabled() {
		return nil
	}
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("kafka message key is empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	body, err := morentevents.MarshalEnvelope(uuid.NewString(), eventType, data)
	if err != nil {
		return err
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strings.ToLower(strings.TrimSpace(key))),
		Value: body,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		p.log.Warn("kafka publish failed", "event_type", eventType, "error", err)
		return err
	}
	p.log.Info("kafka event published", "event_type", eventType, "topic", p.topic)
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
