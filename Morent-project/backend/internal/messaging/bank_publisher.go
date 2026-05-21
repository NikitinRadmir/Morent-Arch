package messaging

import (
	"context"
	"time"
)

// исходящий канал в банковский контур (Kafka). Morent не хранит клиентов/сессии/транзакции.
type BankGateway interface {
	Enabled() bool
	Publish(ctx context.Context, cmd BankCommand) error
}

// команда в банк (топик и схема будут зафиксированы при подключении Kafka).
type BankCommand struct {
	Type      BankCommandType
	RequestID string
	Phone     string
	Password  string
	DisplayName string
	Token     string
	Amount    float64
	RecipientPhone string
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
	BankCmdTransfer     BankCommandType = "bank.transfer"
	BankCmdTransactions BankCommandType = "bank.transactions"
)

type BankNoopGateway struct{}

func (BankNoopGateway) Enabled() bool { return false }

func (BankNoopGateway) Publish(context.Context, BankCommand) error {
	return nil
}
