package postgres

import (
	"time"

	"gorm.io/gorm"
)

type AccountRow struct {
	ID           string `gorm:"primaryKey;size:36"`
	Owner        string `gorm:"size:128;not null"`
	Currency     string `gorm:"size:8;not null"`
	Balance      int64  `gorm:"not null"`
	Status       string `gorm:"size:16;not null"`
	Version      int64  `gorm:"not null"`
	DailyLimit   int64  `gorm:"not null"`
	MonthlyLimit int64  `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ClosedAt     *time.Time
}

func (AccountRow) TableName() string { return "accounts" }

type BankClientRow struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	AccountID    string `gorm:"size:36;index;not null"`
	Phone        string `gorm:"size:16;uniqueIndex;not null"`
	DisplayName  string `gorm:"size:128;not null"`
	PasswordHash string `gorm:"size:255;not null"`
	CardNumber   string `gorm:"size:19;uniqueIndex"`
	ExpDate      string `gorm:"size:8"`
	CVV          string `gorm:"size:4"`
	CardHolder   string `gorm:"size:128"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (BankClientRow) TableName() string { return "bank_clients" }

type BankSessionRow struct {
	Token     string `gorm:"primaryKey;size:64"`
	AccountID string `gorm:"size:36;index;not null"`
	CreatedAt time.Time
}

func (BankSessionRow) TableName() string { return "bank_sessions" }

type TransferRow struct {
	ID             string `gorm:"primaryKey;size:36"`
	FromAccountID  string `gorm:"size:36;index;not null"`
	ToAccountID    string `gorm:"size:36;index;not null"`
	FromCardNumber string `gorm:"size:19"`
	ToCardNumber   string `gorm:"size:19"`
	Amount         int64  `gorm:"not null"`
	Fee            int64  `gorm:"not null"`
	Currency       string `gorm:"size:8;not null"`
	Status         string `gorm:"size:16;not null"`
	CreatedAt      time.Time
	PostedAt       *time.Time
	ReversedAt     *time.Time
}

func (TransferRow) TableName() string { return "transfers" }

type PaymentRow struct {
	ID          string `gorm:"primaryKey;size:36"`
	ReferenceID string `gorm:"size:64;index;not null"`
	AccountID   string `gorm:"size:36;index"`
	CardNumber  string `gorm:"size:19"`
	UserID      string `gorm:"size:64;index;not null"`
	CarID       string `gorm:"size:64;not null"`
	Amount      int64  `gorm:"not null"`
	Currency    string `gorm:"size:8;not null"`
	Status      string `gorm:"size:16;not null"`
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

func (PaymentRow) TableName() string { return "payments" }

type LedgerEntryRow struct {
	ID            string `gorm:"primaryKey;size:36"`
	AccountID     string `gorm:"size:36;index;not null"`
	TransferID    string `gorm:"size:36;index"`
	PaymentID     string `gorm:"size:36;index"`
	OperationType string `gorm:"size:16;not null"`
	Amount        int64  `gorm:"not null"`
	Description   string `gorm:"size:512"`
	CreatedAt     time.Time
}

func (LedgerEntryRow) TableName() string { return "ledger_entries" }

type IdempotencyRow struct {
	Key          string `gorm:"primaryKey;size:128"`
	ResourceID   string `gorm:"size:36;not null"`
	ResourceType string `gorm:"size:32;not null"`
	Fingerprint  string `gorm:"size:64;not null"`
}

func (IdempotencyRow) TableName() string { return "idempotency_records" }

type OperationLimitRow struct {
	Key         string `gorm:"primaryKey;size:96"`
	AccountID   string `gorm:"size:36;index;not null"`
	Period      string `gorm:"size:16;not null"`
	SpentAmount int64  `gorm:"not null"`
	PeriodStart time.Time
	PeriodEnd   time.Time
}

func (OperationLimitRow) TableName() string { return "operation_limits" }

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&AccountRow{},
		&BankClientRow{},
		&BankSessionRow{},
		&TransferRow{},
		&PaymentRow{},
		&LedgerEntryRow{},
		&IdempotencyRow{},
		&OperationLimitRow{},
	)
}
