package kafka

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"user-system/app/service"
)

type Consumer struct {
	reader *kafka.Reader
	sync   *service.MorentSyncService
}

func NewConsumer(brokers, topic, groupID string, sync *service.MorentSyncService) *Consumer {
	brokers = strings.TrimSpace(brokers)
	if topic == "" {
		topic = "morent.users"
	}
	if groupID == "" {
		groupID = "user-system-morent-sync"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        strings.Split(brokers, ","),
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       1e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{reader: reader, sync: sync}
}

func (c *Consumer) Run(ctx context.Context) error {
	slog.Info("kafka consumer started", "topic", c.reader.Config().Topic)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		slog.Info("kafka consumer shutdown signal received")
		_ = c.reader.Close()
	}()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			slog.Error("kafka fetch error", "error", err)
			time.Sleep(time.Second)
			continue
		}

		if err := c.sync.Handle(msg.Value); err != nil {
			slog.Error("kafka event handle error", "offset", msg.Offset, "error", err)
			// не коммитим — повторная обработка
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			slog.Error("kafka commit error", "error", err)
		}
	}
}

func (c *Consumer) Close() error {
	if c.reader == nil {
		return nil
	}
	return c.reader.Close()
}
