// Одноразовая публикация тестового morent.email.send в Kafka (проверка интеграции).
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	morentevents "morent-events"
)

func main() {
	brokers := env("KAFKA_BROKERS", "localhost:9092")
	topic := env("KAFKA_TOPIC_EMAILS", morentevents.TopicEmails)
	to := env("TEST_EMAIL_TO", "verify@morent.local")

	correlationID := strings.ReplaceAll(uuid.NewString(), "-", "")
	payload := morentevents.EmailSendRequested{
		CorrelationID: correlationID,
		To:            to,
		TemplateKey:   morentevents.EmailTemplateBookingConfirmation,
		Variables: map[string]string{
			"user_name":   "Verify User",
			"car_name":    "Test Car",
			"start_date":  "2026-05-22",
			"end_date":    "2026-05-25",
			"total_price": "1500.00",
			"rental_id":   "999",
		},
		Priority: "Normal",
	}

	body, err := morentevents.MarshalEnvelope(uuid.NewString(), morentevents.EventEmailSend, payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
		os.Exit(1)
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP(strings.Split(brokers, ",")...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer w.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(to),
		Value: body,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "publish: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("published morent.email.send topic=%s correlationId=%s to=%s\n", topic, correlationID, to)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
