package repository

import (
	"time"

	"morent-arch/payment-service/internal/domain"
)

type AccountRepository interface {
	CreateAccount(a *domain.Account) error
	GetAccountByID(id string) (*domain.Account, error)
	UpdateAccount(a *domain.Account) error // optimistic lock via Version
	ListAccounts() ([]*domain.Account, error)
}

type TransferRepository interface {
	CreateTransfer(t *domain.Transfer) error
	GetTransferByID(id string) (*domain.Transfer, error)
	UpdateTransfer(t *domain.Transfer) error
	ListTransfers(filter TransferFilter) ([]*domain.Transfer, error)
}

type LedgerRepository interface {
	Append(e *domain.LedgerEntry) error
	ListByAccount(accountID string, limit int, cursor string) (entries []*domain.LedgerEntry, nextCursor string, err error)
}

type PaymentRepository interface {
	CreatePayment(p *domain.Payment) error
	GetPaymentByID(id string) (*domain.Payment, error)
}

type IdempotencyRepository interface {
	Get(key string) (IdempotencyRecord, bool)
	Put(record IdempotencyRecord)
}

type IdempotencyRecord struct {
	Key          string
	ResourceID   string
	ResourceType string
	Fingerprint  string
}

type TransferFilter struct {
	FromAccountID string
	ToAccountID   string
	Status        domain.TransferStatus
	FromDate      *time.Time
	ToDate        *time.Time
}

type LimitsRepository interface {
	GetOperationLimits(accountID string, period domain.LimitPeriod, date time.Time) (*domain.OperationLimits, error)
	RecordOperation(accountID string, period domain.LimitPeriod, amount domain.Money, date time.Time) error
}
