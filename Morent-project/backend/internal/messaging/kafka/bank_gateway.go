package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	morentevents "morent-events"
	"morent-backend/internal/messaging"

	"github.com/segmentio/kafka-go"
)

const bankRequestTimeout = 15 * time.Second

type BankGateway struct {
	cmdWriter    *kafka.Writer
	respReader   *kafka.Reader
	pending      sync.Map
	requestTimeout time.Duration
	log          *slog.Logger
}

func NewBankGateway(brokers, commandsTopic, responsesTopic, groupID string, log *slog.Logger) (*BankGateway, error) {
	brokers = strings.TrimSpace(brokers)
	if brokers == "" {
		return nil, fmt.Errorf("kafka brokers are empty")
	}
	if commandsTopic == "" {
		commandsTopic = morentevents.TopicBankCommands
	}
	if responsesTopic == "" {
		responsesTopic = morentevents.TopicBankResponses
	}
	if groupID == "" {
		groupID = "morent-backend-bank"
	}
	if log == nil {
		log = slog.Default()
	}

	brokerList := strings.Split(brokers, ",")

	cmdWriter := &kafka.Writer{
		Addr:                   kafka.TCP(brokerList...),
		Topic:                  commandsTopic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: true,
	}

	respReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokerList,
		GroupID:        groupID,
		Topic:          responsesTopic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})

	return &BankGateway{
		cmdWriter:      cmdWriter,
		respReader:     respReader,
		requestTimeout: bankRequestTimeout,
		log:            log,
	}, nil
}

func (g *BankGateway) Enabled() bool {
	return g != nil && g.cmdWriter != nil && g.respReader != nil
}

func (g *BankGateway) Close() error {
	var err error
	if g.cmdWriter != nil {
		if e := g.cmdWriter.Close(); e != nil {
			err = e
		}
	}
	if g.respReader != nil {
		if e := g.respReader.Close(); e != nil {
			err = e
		}
	}
	return err
}

func (g *BankGateway) Run(ctx context.Context) error {
	g.log.Info("morent bank kafka response consumer started")
	for {
		msg, err := g.respReader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			g.log.Warn("bank response fetch failed", "error", err)
			continue
		}

		raw, err := morentevents.UnmarshalBankResponse(msg.Value)
		if err != nil {
			g.log.Warn("invalid bank response", "error", err)
			_ = g.respReader.CommitMessages(ctx, msg)
			continue
		}

		g.deliver(messaging.MapBankResponse(raw))

		if err := g.respReader.CommitMessages(ctx, msg); err != nil {
			g.log.Warn("bank response commit failed", "error", err)
		}
	}
}

func (g *BankGateway) deliver(resp messaging.BankResponse) {
	if resp.RequestID == "" {
		return
	}
	v, ok := g.pending.Load(resp.RequestID)
	if !ok {
		return
	}
	ch, ok := v.(chan messaging.BankResponse)
	if !ok {
		return
	}
	select {
	case ch <- resp:
	default:
	}
}

func (g *BankGateway) Request(ctx context.Context, cmd messaging.BankCommand) (messaging.BankResponse, error) {
	if !g.Enabled() {
		return messaging.BankResponse{}, fmt.Errorf("bank kafka gateway is disabled")
	}
	if cmd.RequestID == "" {
		return messaging.BankResponse{}, fmt.Errorf("bank command request_id is empty")
	}

	ch := make(chan messaging.BankResponse, 1)
	g.pending.Store(cmd.RequestID, ch)
	defer g.pending.Delete(cmd.RequestID)

	payload := morentevents.BankCommand{
		Type:           string(cmd.Type),
		RequestID:      cmd.RequestID,
		Phone:          cmd.Phone,
		Password:       cmd.Password,
		DisplayName:    cmd.DisplayName,
		Token:          cmd.Token,
		Amount:          cmd.Amount,
		IdempotencyKey:  cmd.IdempotencyKey,
		RecipientPhone:      cmd.RecipientPhone,
		RecipientCardNumber: cmd.RecipientCardNumber,
		Limit:          cmd.Limit,
		SentAt:         cmd.SentAt,
	}
	if payload.SentAt.IsZero() {
		payload.SentAt = time.Now().UTC()
	}

	body, err := morentevents.MarshalBankCommand(payload)
	if err != nil {
		return messaging.BankResponse{}, err
	}

	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	err = g.cmdWriter.WriteMessages(publishCtx, kafka.Message{
		Key:   []byte(cmd.RequestID),
		Value: body,
		Time:  time.Now().UTC(),
	})
	cancel()
	if err != nil {
		return messaging.BankResponse{}, err
	}

	waitCtx, waitCancel := context.WithTimeout(ctx, g.requestTimeout)
	defer waitCancel()

	select {
	case resp := <-ch:
		return resp, nil
	case <-waitCtx.Done():
		return messaging.BankResponse{RequestID: cmd.RequestID, OK: false, Error: "bank request timed out"}, waitCtx.Err()
	}
}
