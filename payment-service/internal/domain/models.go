package domain

import (
	"errors"
	"time"
)

// Money хранится в минорных единицах (копейки/центы). Знак в LedgerEntry.
type Money = int64

var (
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrCurrencyMismatch        = errors.New("currency mismatch")
	ErrSameAccount             = errors.New("same account")
	ErrNotFound                = errors.New("not found")
	ErrConflict                = errors.New("conflict")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrAccountClosed           = errors.New("account closed")
	ErrUnsupportedCurrency     = errors.New("unsupported currency")
	ErrInvalidOwner            = errors.New("invalid owner")
	ErrTransferNotPosted       = errors.New("transfer not posted")
	ErrTransferAlreadyReversed = errors.New("transfer already reversed")
	ErrLimitExceeded           = errors.New("limit exceeded")
	ErrInvalidFee              = errors.New("invalid fee")
	ErrIdempotencyKeyConflict  = errors.New("idempotency key conflict")
	ErrInvalidPayment          = errors.New("invalid payment")
)

type AccountStatus string

const (
	AccountActive AccountStatus = "active"
	AccountClosed AccountStatus = "closed"
)

type Account struct {
	ID        string
	Owner     string
	Currency  string
	Balance   Money
	Status    AccountStatus
	Version   int64
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	// Лимиты на операции
	DailyLimit   Money // дневной лимит на списания
	MonthlyLimit Money // месячный лимит на списания
}

type TransferStatus string

const (
	TransferPending  TransferStatus = "pending"
	TransferPosted   TransferStatus = "posted"
	TransferReversed TransferStatus = "reversed"
)

type Transfer struct {
	ID            string
	FromAccountID string
	ToAccountID   string
	Amount        Money
	Fee           Money // комиссия, списывается с отправителя
	Currency      string
	Status        TransferStatus
	CreatedAt     time.Time
	PostedAt      *time.Time
	ReversedAt    *time.Time
}

type PaymentStatus string

const (
	PaymentPending   PaymentStatus = "pending"
	PaymentSucceeded PaymentStatus = "succeeded"
	PaymentFailed    PaymentStatus = "failed"
)

type Payment struct {
	ID          string
	ReferenceID string
	UserID      string
	CarID       string
	Amount      Money
	Currency    string
	Status      PaymentStatus
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

type OperationType string

const (
	OperationDeposit  OperationType = "deposit"
	OperationWithdraw OperationType = "withdraw"
	OperationTransfer OperationType = "transfer"
	OperationFee      OperationType = "fee"
	OperationReversal OperationType = "reversal"
)

type LedgerEntry struct {
	ID            string
	AccountID     string
	TransferID    string
	OperationType OperationType
	Amount        Money // +credit, -debit
	Description   string
	CreatedAt     time.Time
}
