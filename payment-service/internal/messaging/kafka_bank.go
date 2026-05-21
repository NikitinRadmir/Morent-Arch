package messaging

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"morent-arch/payment-service/internal/bank"
	"morent-arch/payment-service/internal/config"
	"morent-arch/payment-service/internal/repository/memory"
	"morent-arch/payment-service/internal/service"
	morentevents "morent-events"

	"github.com/segmentio/kafka-go"
)

type BankKafka struct {
	processor *bank.Processor
	reader    *kafka.Reader
	writer    *kafka.Writer
	log       *slog.Logger
}

func NewBankKafka(cfg config.Config, log *slog.Logger) (*BankKafka, error) {
	if log == nil {
		log = slog.Default()
	}
	brokers := strings.Split(cfg.KafkaBrokers, ",")
	store := memory.NewStore()
	pay := service.NewPaymentService(store, store, store, store, store, store)
	proc := bank.NewProcessor(pay, bank.NewRegistry())

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        cfg.KafkaGroupPaymentBank,
		Topic:          cfg.KafkaTopicBankCommands,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  cfg.KafkaTopicBankResponses,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}

	return &BankKafka{
		processor: proc,
		reader:    reader,
		writer:    writer,
		log:       log,
	}, nil
}

func (k *BankKafka) Run(ctx context.Context) error {
	k.log.Info("payment-service bank kafka consumer started")
	for {
		msg, err := k.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			k.log.Warn("kafka fetch failed", "error", err)
			continue
		}

		cmd, err := morentevents.UnmarshalBankCommand(msg.Value)
		if err != nil {
			k.log.Warn("invalid bank command", "error", err)
			_ = k.reader.CommitMessages(ctx, msg)
			continue
		}

		resp := k.processor.Handle(ctx, cmd)
		body, err := morentevents.MarshalBankResponse(resp)
		if err != nil {
			k.log.Warn("marshal bank response failed", "error", err)
			_ = k.reader.CommitMessages(ctx, msg)
			continue
		}

		publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = k.writer.WriteMessages(publishCtx, kafka.Message{
			Key:   []byte(resp.RequestID),
			Value: body,
			Time:  time.Now().UTC(),
		})
		cancel()
		if err != nil {
			k.log.Warn("kafka publish response failed", "request_id", resp.RequestID, "error", err)
			continue
		}

		if err := k.reader.CommitMessages(ctx, msg); err != nil {
			k.log.Warn("kafka commit failed", "error", err)
		}
	}
}

func (k *BankKafka) Close() error {
	var err error
	if k.reader != nil {
		if e := k.reader.Close(); e != nil {
			err = e
		}
	}
	if k.writer != nil {
		if e := k.writer.Close(); e != nil {
			err = e
		}
	}
	return err
}
