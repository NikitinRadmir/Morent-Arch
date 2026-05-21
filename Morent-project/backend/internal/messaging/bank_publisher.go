package messaging

import (
	"context"
	"time"
)

// BankGateway — request-reply в payment-service через Kafka.
type BankGateway interface {
	Enabled() bool
	Request(ctx context.Context, cmd BankCommand) (BankResponse, error)
}

// команда в банк (топик и схема будут зафиксированы при подключении Kafka).
type BankCommand struct {
	Type      BankCommandType
	RequestID string
	Phone     string
	Password  string
	DisplayName string
	Token     string
	Amount              float64
	IdempotencyKey      string
	RecipientPhone      string
	RecipientCardNumber string
	Limit     int
	SentAt    time.Time
}

type BankCommandType string

const (
	BankCmdRegister     BankCommandType = "bank.register"
	BankCmdLogin        BankCommandType = "bank.login"
	BankCmdLogout       BankCommandType = "bank.logout"
	BankCmdGetProfile   BankCommandType = "bank.profile"
	BankCmdDeposit      BankCommandType = "bank.deposit"
	BankCmdPay          BankCommandType = "bank.pay"
	BankCmdTransfer     BankCommandType = "bank.transfer"
	BankCmdTransactions BankCommandType = "bank.transactions"
)

type BankNoopGateway struct{}

func (BankNoopGateway) Enabled() bool { return false }

func (BankNoopGateway) Request(context.Context, BankCommand) (BankResponse, error) {
	return BankResponse{}, nil
}
