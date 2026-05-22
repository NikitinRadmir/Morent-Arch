package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"morent-arch/payment-service/internal/domain"
	"morent-arch/payment-service/internal/observability"
	"morent-arch/payment-service/internal/repository"
)

type PaymentService struct {
	mu          sync.Mutex
	accounts    repository.AccountRepository
	transfers   repository.TransferRepository
	ledger      repository.LedgerRepository
	payments    repository.PaymentRepository
	idempoStore repository.IdempotencyRepository
	limits      repository.LimitsRepository
}

func NewPaymentService(a repository.AccountRepository, t repository.TransferRepository, l repository.LedgerRepository, p repository.PaymentRepository, id repository.IdempotencyRepository, lim repository.LimitsRepository) *PaymentService {
	return &PaymentService{accounts: a, transfers: t, ledger: l, payments: p, idempoStore: id, limits: lim}
}

const (
	idempotencyResourceTransfer = "transfer"
	idempotencyResourcePayment  = "payment"
)

type CreateAccountInput struct {
	Owner        string
	Currency     string
	DailyLimit   domain.Money
	MonthlyLimit domain.Money
}

type TransferInput struct {
	FromAccountID  string
	ToAccountID    string
	FromCardNumber string
	ToCardNumber   string
	Amount         domain.Money
	Fee            domain.Money // комиссия (опционально, 0 = без комиссии)
	Currency       string
	IdempotencyKey string
}

type BalanceChangeInput struct {
	AccountID string
	Amount    domain.Money
	Currency  string
}

type CreatePaymentInput struct {
	ReferenceID    string
	AccountID      string
	CardNumber     string
	UserID         string
	CarID          string
	Amount         domain.Money
	Currency       string
	IdempotencyKey string
}

type ListAccountsFilter struct {
	Status domain.AccountStatus
	Owner  string
}

// AccountSummary агрегирует базовую аналитику по одному счёту.
type AccountSummary struct {
	Account           *domain.Account
	TotalDeposits     domain.Money
	TotalWithdraws    domain.Money
	TotalInTransfers  domain.Money
	TotalOutTransfers domain.Money
	TotalFees         domain.Money
}

func (s *PaymentService) CreateAccount(ctx context.Context, in CreateAccountInput) (*domain.Account, error) {
	if strings.TrimSpace(in.Owner) == "" {
		return nil, domain.ErrInvalidOwner
	}
	if !domain.IsSupportedCurrency(in.Currency) {
		return nil, domain.ErrUnsupportedCurrency
	}

	now := time.Now()

	acc := &domain.Account{
		ID:           newID(),
		Owner:        strings.TrimSpace(in.Owner),
		Currency:     in.Currency,
		Balance:      0,
		Status:       domain.AccountActive,
		Version:      1,
		DailyLimit:   in.DailyLimit,
		MonthlyLimit: in.MonthlyLimit,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.accounts.CreateAccount(acc); err != nil {
		return nil, err
	}
	logInfo(ctx, "account_created",
		"account_id", acc.ID,
		"owner", acc.Owner,
		"currency", acc.Currency,
		"daily_limit_minor", acc.DailyLimit,
		"monthly_limit_minor", acc.MonthlyLimit,
	)
	return acc, nil
}

func (s *PaymentService) GetAccount(_ context.Context, id string) (*domain.Account, error) {
	return s.accounts.GetAccountByID(id)
}

func (s *PaymentService) ListAccounts(_ context.Context, filter ListAccountsFilter) ([]*domain.Account, error) {
	items, err := s.accounts.ListAccounts()
	if err != nil {
		return nil, err
	}
	var out []*domain.Account
	for _, acc := range items {
		if filter.Status != "" && acc.Status != filter.Status {
			continue
		}
		if filter.Owner != "" && !strings.Contains(strings.ToLower(acc.Owner), strings.ToLower(filter.Owner)) {
			continue
		}
		out = append(out, acc)
	}
	return out, nil
}

func (s *PaymentService) GetTransfer(_ context.Context, id string) (*domain.Transfer, error) {
	return s.transfers.GetTransferByID(id)
}

func (s *PaymentService) CreatePayment(ctx context.Context, in CreatePaymentInput) (*domain.Payment, error) {
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	referenceID := strings.TrimSpace(in.ReferenceID)
	userID := strings.TrimSpace(in.UserID)
	carID := strings.TrimSpace(in.CarID)
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if referenceID == "" || userID == "" || carID == "" {
		return nil, domain.ErrInvalidPayment
	}
	if !domain.IsSupportedCurrency(currency) {
		return nil, domain.ErrUnsupportedCurrency
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recordPaymentLocked(ctx, in, referenceID, userID, carID, currency)
}

// RecordPayment сохраняет платёж без блокировки (после Withdraw/Transfer в том же потоке).
func (s *PaymentService) RecordPayment(ctx context.Context, in CreatePaymentInput) (*domain.Payment, error) {
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	referenceID := strings.TrimSpace(in.ReferenceID)
	userID := strings.TrimSpace(in.UserID)
	carID := strings.TrimSpace(in.CarID)
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if referenceID == "" || userID == "" || carID == "" {
		return nil, domain.ErrInvalidPayment
	}
	if !domain.IsSupportedCurrency(currency) {
		return nil, domain.ErrUnsupportedCurrency
	}
	return s.recordPaymentLocked(ctx, in, referenceID, userID, carID, currency)
}

func (s *PaymentService) recordPaymentLocked(ctx context.Context, in CreatePaymentInput, referenceID, userID, carID, currency string) (*domain.Payment, error) {
	input := in
	input.ReferenceID = referenceID
	input.UserID = userID
	input.CarID = carID
	input.Currency = currency
	fingerprint := paymentFingerprint(input)
	if input.IdempotencyKey != "" {
		if record, ok := s.idempoStore.Get(input.IdempotencyKey); ok {
			if record.ResourceType != idempotencyResourcePayment || record.Fingerprint != fingerprint {
				logWarn(ctx, "payment_idempotency_conflict",
					"idempotency_key", input.IdempotencyKey,
					"reference_id", referenceID,
				)
				return nil, domain.ErrIdempotencyKeyConflict
			}
			logInfo(ctx, "payment_idempotency_replayed",
				"idempotency_key", input.IdempotencyKey,
				"payment_id", record.ResourceID,
			)
			return s.payments.GetPaymentByID(record.ResourceID)
		}
	}

	now := time.Now()
	payment := &domain.Payment{
		ID:          newID(),
		ReferenceID: referenceID,
		AccountID:   strings.TrimSpace(input.AccountID),
		CardNumber:  strings.TrimSpace(input.CardNumber),
		UserID:      userID,
		CarID:       carID,
		Amount:      input.Amount,
		Currency:    currency,
		Status:      domain.PaymentSucceeded,
		CreatedAt:   now,
		ProcessedAt: &now,
	}
	if err := s.payments.CreatePayment(payment); err != nil {
		return nil, err
	}
	if input.IdempotencyKey != "" {
		s.idempoStore.Put(repository.IdempotencyRecord{
			Key:          input.IdempotencyKey,
			ResourceID:   payment.ID,
			ResourceType: idempotencyResourcePayment,
			Fingerprint:  fingerprint,
		})
	}
	logInfo(ctx, "payment_succeeded",
		"payment_id", payment.ID,
		"reference_id", payment.ReferenceID,
		"user_id", payment.UserID,
		"car_id", payment.CarID,
		"amount_minor", payment.Amount,
		"currency", payment.Currency,
		"idempotency_key_present", input.IdempotencyKey != "",
	)
	return payment, nil
}

func (s *PaymentService) GetPayment(_ context.Context, id string) (*domain.Payment, error) {
	return s.payments.GetPaymentByID(id)
}

// Transfer с идемпотентностью, комиссиями и лимитами, с оптимистичными ретраями.
func (s *PaymentService) Transfer(ctx context.Context, in TransferInput) (*domain.Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.transferLocked(ctx, in)
}

func (s *PaymentService) transferLocked(ctx context.Context, in TransferInput) (*domain.Transfer, error) {
	if in.FromAccountID == in.ToAccountID {
		return nil, domain.ErrSameAccount
	}
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if in.Fee < 0 {
		return nil, domain.ErrInvalidFee
	}
	fingerprint := transferFingerprint(in)
	if in.IdempotencyKey != "" {
		if record, ok := s.idempoStore.Get(in.IdempotencyKey); ok {
			if record.ResourceType != idempotencyResourceTransfer || record.Fingerprint != fingerprint {
				logWarn(ctx, "transfer_idempotency_conflict",
					"idempotency_key", in.IdempotencyKey,
					"from_account_id", in.FromAccountID,
					"to_account_id", in.ToAccountID,
				)
				return nil, domain.ErrIdempotencyKeyConflict
			}
			logInfo(ctx, "transfer_idempotency_replayed",
				"idempotency_key", in.IdempotencyKey,
				"transfer_id", record.ResourceID,
			)
			return s.transfers.GetTransferByID(record.ResourceID)
		}
	}

	from, err := s.accounts.GetAccountByID(in.FromAccountID)
	if err != nil {
		return nil, err
	}
	if from.Status == domain.AccountClosed {
		return nil, domain.ErrAccountClosed
	}
	to, err := s.accounts.GetAccountByID(in.ToAccountID)
	if err != nil {
		return nil, err
	}
	if to.Status == domain.AccountClosed {
		return nil, domain.ErrAccountClosed
	}
	if from.Currency != in.Currency || to.Currency != in.Currency {
		return nil, domain.ErrCurrencyMismatch
	}

	totalDebit := in.Amount + in.Fee
	if from.Balance < totalDebit {
		return nil, domain.ErrInsufficientFunds
	}

	// Проверка лимитов
	now := time.Now()
	if from.DailyLimit > 0 {
		dailyLimits, err := s.limits.GetOperationLimits(from.ID, domain.LimitDaily, now)
		if err == nil && !dailyLimits.CheckLimit(from.DailyLimit, totalDebit) {
			return nil, domain.ErrLimitExceeded
		}
	}
	if from.MonthlyLimit > 0 {
		monthlyLimits, err := s.limits.GetOperationLimits(from.ID, domain.LimitMonthly, now)
		if err == nil && !monthlyLimits.CheckLimit(from.MonthlyLimit, totalDebit) {
			return nil, domain.ErrLimitExceeded
		}
	}

	tr := &domain.Transfer{
		ID:             newID(),
		FromAccountID:  from.ID,
		ToAccountID:    to.ID,
		FromCardNumber: strings.TrimSpace(in.FromCardNumber),
		ToCardNumber:   strings.TrimSpace(in.ToCardNumber),
		Amount:         in.Amount,
		Fee:            in.Fee,
		Currency:       in.Currency,
		Status:         domain.TransferPending,
		CreatedAt:      now,
	}
	if err := s.transfers.CreateTransfer(tr); err != nil {
		return nil, err
	}

	for retries := 0; retries < 3; retries++ {
		// перезагружаем последние версии
		from, err = s.accounts.GetAccountByID(in.FromAccountID)
		if err != nil {
			return nil, err
		}
		to, err = s.accounts.GetAccountByID(in.ToAccountID)
		if err != nil {
			return nil, err
		}
		if from.Balance < totalDebit {
			return nil, domain.ErrInsufficientFunds
		}

		from.Balance -= totalDebit
		to.Balance += in.Amount

		if err = s.accounts.UpdateAccount(from); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}
		if err = s.accounts.UpdateAccount(to); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}

		entryTime := time.Now()
		_ = s.ledger.Append(&domain.LedgerEntry{
			ID:            newID(),
			AccountID:     from.ID,
			TransferID:    tr.ID,
			OperationType: domain.OperationTransfer,
			Amount:        -in.Amount,
			Description:   "Transfer to " + to.ID,
			CreatedAt:     entryTime,
		})
		if in.Fee > 0 {
			_ = s.ledger.Append(&domain.LedgerEntry{
				ID:            newID(),
				AccountID:     from.ID,
				TransferID:    tr.ID,
				OperationType: domain.OperationFee,
				Amount:        -in.Fee,
				Description:   "Transfer fee",
				CreatedAt:     entryTime,
			})
		}
		_ = s.ledger.Append(&domain.LedgerEntry{
			ID:            newID(),
			AccountID:     to.ID,
			TransferID:    tr.ID,
			OperationType: domain.OperationTransfer,
			Amount:        in.Amount,
			Description:   "Transfer from " + from.ID,
			CreatedAt:     entryTime,
		})

		// Записываем операцию в лимиты
		_ = s.limits.RecordOperation(from.ID, domain.LimitDaily, totalDebit, now)
		_ = s.limits.RecordOperation(from.ID, domain.LimitMonthly, totalDebit, now)

		tr.Status = domain.TransferPosted
		tr.PostedAt = &now
		// Сохраняем обновлённый статус перевода в хранилище
		if err = s.transfers.UpdateTransfer(tr); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}
		break
	}
	if tr.Status != domain.TransferPosted {
		return nil, domain.ErrConflict
	}
	if in.IdempotencyKey != "" {
		s.idempoStore.Put(repository.IdempotencyRecord{
			Key:          in.IdempotencyKey,
			ResourceID:   tr.ID,
			ResourceType: idempotencyResourceTransfer,
			Fingerprint:  fingerprint,
		})
	}
	logInfo(ctx, "transfer_posted",
		"transfer_id", tr.ID,
		"from_account_id", tr.FromAccountID,
		"to_account_id", tr.ToAccountID,
		"amount_minor", tr.Amount,
		"fee_minor", tr.Fee,
		"currency", tr.Currency,
		"idempotency_key_present", in.IdempotencyKey != "",
	)
	return tr, nil
}

func (s *PaymentService) ListHistory(_ context.Context, accountID string, limit int, cursor string) ([]*domain.LedgerEntry, string, error) {
	if _, err := s.accounts.GetAccountByID(accountID); err != nil {
		return nil, "", err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.ledger.ListByAccount(accountID, limit, cursor)
}

// GetAccountSummary возвращает агрегированную сводку по одному счёту:
// баланс, сумму депозитов/снятий и входящих/исходящих переводов.
func (s *PaymentService) GetAccountSummary(_ context.Context, accountID string) (*AccountSummary, error) {
	acc, err := s.accounts.GetAccountByID(accountID)
	if err != nil {
		return nil, err
	}

	entries, _, err := s.ledger.ListByAccount(accountID, 1000, "")
	if err != nil {
		return nil, err
	}

	sum := &AccountSummary{
		Account: acc,
	}
	for _, e := range entries {
		switch e.OperationType {
		case domain.OperationDeposit:
			sum.TotalDeposits += e.Amount
		case domain.OperationWithdraw:
			sum.TotalWithdraws += -e.Amount
		case domain.OperationTransfer:
			if e.Amount > 0 {
				sum.TotalInTransfers += e.Amount
			} else {
				sum.TotalOutTransfers += -e.Amount
			}
		case domain.OperationFee:
			sum.TotalFees += -e.Amount
		}
	}
	return sum, nil
}

func (s *PaymentService) Deposit(ctx context.Context, in BalanceChangeInput) (*domain.Account, error) {
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.changeBalance(ctx, in.AccountID, in.Currency, in.Amount)
}

func (s *PaymentService) Withdraw(ctx context.Context, in BalanceChangeInput) (*domain.Account, error) {
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.changeBalance(ctx, in.AccountID, in.Currency, -in.Amount)
}

func (s *PaymentService) CloseAccount(_ context.Context, accountID string) (*domain.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for retries := 0; retries < 3; retries++ {
		acc, err := s.accounts.GetAccountByID(accountID)
		if err != nil {
			return nil, err
		}
		if acc.Status == domain.AccountClosed {
			return acc, nil
		}
		if acc.Balance != 0 {
			return nil, domain.ErrConflict
		}
		now := time.Now()
		acc.Status = domain.AccountClosed
		acc.ClosedAt = &now
		if err := s.accounts.UpdateAccount(acc); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}
		return acc, nil
	}
	return nil, domain.ErrConflict
}

func (s *PaymentService) changeBalance(ctx context.Context, accountID, currency string, delta domain.Money) (*domain.Account, error) {
	for retries := 0; retries < 3; retries++ {
		acc, err := s.accounts.GetAccountByID(accountID)
		if err != nil {
			return nil, err
		}
		if acc.Status == domain.AccountClosed {
			return nil, domain.ErrAccountClosed
		}
		if acc.Currency != currency {
			return nil, domain.ErrCurrencyMismatch
		}
		if delta < 0 && acc.Balance < -delta {
			return nil, domain.ErrInsufficientFunds
		}
		acc.Balance += delta
		if err := s.accounts.UpdateAccount(acc); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}
		amount := delta
		opType := domain.OperationDeposit
		desc := "Deposit"
		if delta < 0 {
			opType = domain.OperationWithdraw
			desc = "Withdrawal"
		}
		entry := &domain.LedgerEntry{
			ID:            newID(),
			AccountID:     accountID,
			OperationType: opType,
			Amount:        amount,
			Description:   desc,
			CreatedAt:     time.Now(),
		}
		_ = s.ledger.Append(entry)
		logInfo(ctx, "balance_changed",
			"account_id", accountID,
			"operation_type", opType,
			"amount_minor", amount,
			"currency", currency,
			"balance_minor", acc.Balance,
		)
		return acc, nil
	}
	return nil, domain.ErrConflict
}

// ReverseTransfer отменяет выполненный перевод, возвращая средства обратно.
func (s *PaymentService) ReverseTransfer(ctx context.Context, transferID string) (*domain.Transfer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tr, err := s.transfers.GetTransferByID(transferID)
	if err != nil {
		return nil, err
	}
	if tr.Status == domain.TransferReversed {
		return nil, domain.ErrTransferAlreadyReversed
	}
	if tr.Status != domain.TransferPosted {
		return nil, domain.ErrTransferNotPosted
	}

	// создаем обратный перевод (возвращаем сумму + комиссию)
	reverseTr := &domain.Transfer{
		ID:            newID(),
		FromAccountID: tr.ToAccountID,
		ToAccountID:   tr.FromAccountID,
		Amount:        tr.Amount,
		Fee:           0,
		Currency:      tr.Currency,
		Status:        domain.TransferPending,
		CreatedAt:     time.Now(),
	}
	if err := s.transfers.CreateTransfer(reverseTr); err != nil {
		return nil, err
	}

	// выполняем обратный перевод с оптимистичными ретраями
	for retries := 0; retries < 3; retries++ {
		from, err := s.accounts.GetAccountByID(reverseTr.FromAccountID) // это получатель оригинального перевода (to)
		if err != nil {
			return nil, err
		}
		to, err := s.accounts.GetAccountByID(reverseTr.ToAccountID) // это отправитель оригинального перевода (from)
		if err != nil {
			return nil, err
		}
		// При реверсе: получатель возвращает только сумму перевода (без комиссии)
		// Отправитель получает обратно сумму + комиссию
		if from.Balance < tr.Amount {
			return nil, domain.ErrInsufficientFunds
		}

		// С получателя списываем только сумму перевода
		from.Balance -= tr.Amount
		// Отправителю возвращаем сумму перевода; комиссия остается уже учтенной как fee.
		to.Balance += tr.Amount

		if err = s.accounts.UpdateAccount(from); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}
		if err = s.accounts.UpdateAccount(to); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}

		entryTime := time.Now()
		// Запись для получателя: списываем сумму перевода
		_ = s.ledger.Append(&domain.LedgerEntry{
			ID:            newID(),
			AccountID:     from.ID,
			TransferID:    reverseTr.ID,
			OperationType: domain.OperationReversal,
			Amount:        -tr.Amount,
			Description:   "Reversal of transfer " + transferID + " (return amount)",
			CreatedAt:     entryTime,
		})
		// Запись для отправителя: возвращаем сумму перевода
		_ = s.ledger.Append(&domain.LedgerEntry{
			ID:            newID(),
			AccountID:     to.ID,
			TransferID:    reverseTr.ID,
			OperationType: domain.OperationReversal,
			Amount:        tr.Amount,
			Description:   "Reversal of transfer " + transferID + " (return amount)",
			CreatedAt:     entryTime,
		})
		now := time.Now()
		reverseTr.Status = domain.TransferPosted
		reverseTr.PostedAt = &now
		// Сохраняем статус обратного перевода
		if err = s.transfers.UpdateTransfer(reverseTr); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}

		// Помечаем оригинальный перевод как отмененный
		tr.Status = domain.TransferReversed
		tr.ReversedAt = &now
		if err = s.transfers.UpdateTransfer(tr); err != nil {
			if err == domain.ErrConflict {
				continue
			}
			return nil, err
		}
		break
	}
	if reverseTr.Status != domain.TransferPosted {
		return nil, domain.ErrConflict
	}
	logInfo(ctx, "transfer_reversed",
		"original_transfer_id", tr.ID,
		"reverse_transfer_id", reverseTr.ID,
		"from_account_id", reverseTr.FromAccountID,
		"to_account_id", reverseTr.ToAccountID,
		"amount_minor", reverseTr.Amount,
		"currency", reverseTr.Currency,
	)
	return reverseTr, nil
}

// GetOwnerBalance возвращает суммарный баланс всех счетов владельца по валютам.
type OwnerBalance struct {
	Owner    string
	Balances map[string]domain.Money // currency -> balance
}

func (s *PaymentService) GetOwnerBalance(_ context.Context, owner string) (*OwnerBalance, error) {
	accounts, err := s.accounts.ListAccounts()
	if err != nil {
		return nil, err
	}
	balances := make(map[string]domain.Money)
	for _, acc := range accounts {
		if strings.ToLower(acc.Owner) == strings.ToLower(owner) && acc.Status == domain.AccountActive {
			balances[acc.Currency] += acc.Balance
		}
	}
	return &OwnerBalance{Owner: owner, Balances: balances}, nil
}

// TransferStats содержит статистику по переводам.
type TransferStats struct {
	TotalTransfers    int
	TotalAmount       domain.Money
	PostedTransfers   int
	ReversedTransfers int
	FromDate          *time.Time
	ToDate            *time.Time
}

func (s *PaymentService) GetTransferStats(_ context.Context, filter repository.TransferFilter) (*TransferStats, error) {
	transfers, err := s.transfers.ListTransfers(filter)
	if err != nil {
		return nil, err
	}
	stats := &TransferStats{
		FromDate: filter.FromDate,
		ToDate:   filter.ToDate,
	}
	for _, tr := range transfers {
		stats.TotalTransfers++
		if tr.Status == domain.TransferPosted {
			stats.PostedTransfers++
			stats.TotalAmount += tr.Amount
		} else if tr.Status == domain.TransferReversed {
			stats.ReversedTransfers++
		}
	}
	return stats, nil
}

func (s *PaymentService) ListTransfers(_ context.Context, filter repository.TransferFilter) ([]*domain.Transfer, error) {
	return s.transfers.ListTransfers(filter)
}

// BatchTransferInput содержит данные для одного перевода в батче.
type BatchTransferItem struct {
	FromAccountID string
	ToAccountID   string
	Amount        domain.Money
	Fee           domain.Money
	Currency      string
	Description   string
}

// BatchTransferResult содержит результаты батч-операции.
type BatchTransferResult struct {
	Total      int
	Successful int
	Failed     int
	Transfers  []*domain.Transfer
	Errors     []BatchTransferError
}

type BatchTransferError struct {
	Index int
	Item  BatchTransferItem
	Error string
}

// BatchTransfer выполняет несколько переводов атомарно (все или ничего).
func (s *PaymentService) BatchTransfer(ctx context.Context, items []BatchTransferItem, idempotencyKeyPrefix string) (*BatchTransferResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(items) == 0 {
		return &BatchTransferResult{}, nil
	}

	result := &BatchTransferResult{
		Total:     len(items),
		Transfers: make([]*domain.Transfer, 0, len(items)),
		Errors:    make([]BatchTransferError, 0),
	}
	requiredByAccount := make(map[string]domain.Money)

	// Валидация всех операций перед выполнением
	for i, item := range items {
		if item.FromAccountID == item.ToAccountID {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrSameAccount.Error(),
			})
			continue
		}
		if item.Amount <= 0 {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrInvalidAmount.Error(),
			})
			continue
		}
		if item.Fee < 0 {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrInvalidFee.Error(),
			})
			continue
		}
		if !domain.IsSupportedCurrency(item.Currency) {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrUnsupportedCurrency.Error(),
			})
			continue
		}
		if idempotencyKeyPrefix != "" {
			idemKey := fmt.Sprintf("%s:%d", idempotencyKeyPrefix, i)
			input := TransferInput{
				FromAccountID:  item.FromAccountID,
				ToAccountID:    item.ToAccountID,
				Amount:         item.Amount,
				Fee:            item.Fee,
				Currency:       item.Currency,
				IdempotencyKey: idemKey,
			}
			if record, ok := s.idempoStore.Get(idemKey); ok {
				if record.ResourceType != idempotencyResourceTransfer || record.Fingerprint != transferFingerprint(input) {
					result.Errors = append(result.Errors, BatchTransferError{
						Index: i,
						Item:  item,
						Error: domain.ErrIdempotencyKeyConflict.Error(),
					})
				}
				continue
			}
		}

		// Проверяем существование счетов
		from, err := s.accounts.GetAccountByID(item.FromAccountID)
		if err != nil {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: err.Error(),
			})
			continue
		}
		if from.Status == domain.AccountClosed {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrAccountClosed.Error(),
			})
			continue
		}

		to, err := s.accounts.GetAccountByID(item.ToAccountID)
		if err != nil {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: err.Error(),
			})
			continue
		}
		if to.Status == domain.AccountClosed {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrAccountClosed.Error(),
			})
			continue
		}
		if from.Currency != item.Currency || to.Currency != item.Currency {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrCurrencyMismatch.Error(),
			})
			continue
		}

		totalDebit := item.Amount + item.Fee
		if from.Balance < totalDebit {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrInsufficientFunds.Error(),
			})
			continue
		}
		requiredByAccount[from.ID] += totalDebit
	}

	for accountID, required := range requiredByAccount {
		acc, err := s.accounts.GetAccountByID(accountID)
		if err != nil {
			result.Errors = append(result.Errors, BatchTransferError{Index: -1, Error: err.Error()})
			continue
		}
		if acc.Balance < required {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: -1,
				Error: fmt.Sprintf("%s: %s", accountID, domain.ErrInsufficientFunds.Error()),
			})
		}
		now := time.Now()
		if acc.DailyLimit > 0 {
			dailyLimits, err := s.limits.GetOperationLimits(accountID, domain.LimitDaily, now)
			if err == nil && !dailyLimits.CheckLimit(acc.DailyLimit, required) {
				result.Errors = append(result.Errors, BatchTransferError{
					Index: -1,
					Error: fmt.Sprintf("%s: %s", accountID, domain.ErrLimitExceeded.Error()),
				})
			}
		}
		if acc.MonthlyLimit > 0 {
			monthlyLimits, err := s.limits.GetOperationLimits(accountID, domain.LimitMonthly, now)
			if err == nil && !monthlyLimits.CheckLimit(acc.MonthlyLimit, required) {
				result.Errors = append(result.Errors, BatchTransferError{
					Index: -1,
					Error: fmt.Sprintf("%s: %s", accountID, domain.ErrLimitExceeded.Error()),
				})
			}
		}
	}

	// Если есть ошибки валидации, возвращаем их
	if len(result.Errors) > 0 {
		result.Failed = len(result.Errors)
		return result, nil
	}

	// Выполняем все переводы
	for i, item := range items {
		idemKey := ""
		if idempotencyKeyPrefix != "" {
			idemKey = fmt.Sprintf("%s:%d", idempotencyKeyPrefix, i)
		}

		tr, err := s.transferLocked(ctx, TransferInput{
			FromAccountID:  item.FromAccountID,
			ToAccountID:    item.ToAccountID,
			Amount:         item.Amount,
			Fee:            item.Fee,
			Currency:       item.Currency,
			IdempotencyKey: idemKey,
		})

		if err != nil {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: err.Error(),
			})
			result.Failed++
		} else {
			result.Transfers = append(result.Transfers, tr)
			result.Successful++
		}
	}

	logInfo(ctx, "batch_transfer_completed",
		"total", result.Total,
		"successful", result.Successful,
		"failed", result.Failed,
		"idempotency_key_prefix_present", idempotencyKeyPrefix != "",
	)
	return result, nil
}

func transferFingerprint(in TransferInput) string {
	raw := fmt.Sprintf("%s|%s|%d|%d|%s", in.FromAccountID, in.ToAccountID, in.Amount, in.Fee, in.Currency)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func paymentFingerprint(in CreatePaymentInput) string {
	raw := fmt.Sprintf("%s|%s|%s|%d|%s", in.ReferenceID, in.UserID, in.CarID, in.Amount, in.Currency)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func logInfo(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, withRequestID(ctx, args)...)
}

func logWarn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, withRequestID(ctx, args)...)
}

func withRequestID(ctx context.Context, args []any) []any {
	if requestID := observability.RequestID(ctx); requestID != "" {
		return append([]any{"request_id", requestID}, args...)
	}
	return args
}
