package service

import (
	"context"
	"strings"
	"time"

	"example.com/go-payments/internal/domain"
	"example.com/go-payments/internal/repository"
	"github.com/google/uuid"
)

type PaymentService struct {
	accounts    repository.AccountRepository
	transfers   repository.TransferRepository
	ledger      repository.LedgerRepository
	idempoStore repository.IdempotencyRepository
	limits      repository.LimitsRepository
}

func NewPaymentService(a repository.AccountRepository, t repository.TransferRepository, l repository.LedgerRepository, id repository.IdempotencyRepository, lim repository.LimitsRepository) *PaymentService {
	return &PaymentService{accounts: a, transfers: t, ledger: l, idempoStore: id, limits: lim}
}

type CreateAccountInput struct {
	Owner        string
	Currency     string
	DailyLimit   domain.Money
	MonthlyLimit domain.Money
}

type TransferInput struct {
	FromAccountID  string
	ToAccountID    string
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

func (s *PaymentService) CreateAccount(_ context.Context, in CreateAccountInput) (*domain.Account, error) {
	if strings.TrimSpace(in.Owner) == "" {
		return nil, domain.ErrInvalidOwner
	}
	if !domain.IsSupportedCurrency(in.Currency) {
		return nil, domain.ErrUnsupportedCurrency
	}

	now := time.Now()

	acc := &domain.Account{
		ID:           uuid.NewString(),
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

// Transfer с идемпотентностью, комиссиями и лимитами, с оптимистичными ретраями.
func (s *PaymentService) Transfer(_ context.Context, in TransferInput) (*domain.Transfer, error) {
	if in.FromAccountID == in.ToAccountID {
		return nil, domain.ErrSameAccount
	}
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if in.Fee < 0 {
		return nil, domain.ErrInvalidFee
	}
	if in.IdempotencyKey != "" {
		if tid, ok := s.idempoStore.Get(in.IdempotencyKey); ok {
			return s.transfers.GetTransferByID(tid)
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
		ID:            uuid.NewString(),
		FromAccountID: from.ID,
		ToAccountID:   to.ID,
		Amount:        in.Amount,
		Fee:           in.Fee,
		Currency:      in.Currency,
		Status:        domain.TransferPending,
		CreatedAt:     now,
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
			ID:            uuid.NewString(),
			AccountID:     from.ID,
			TransferID:    tr.ID,
			OperationType: domain.OperationTransfer,
			Amount:        -in.Amount,
			Description:   "Transfer to " + to.ID,
			CreatedAt:     entryTime,
		})
		if in.Fee > 0 {
			_ = s.ledger.Append(&domain.LedgerEntry{
				ID:            uuid.NewString(),
				AccountID:     from.ID,
				TransferID:    tr.ID,
				OperationType: domain.OperationFee,
				Amount:        -in.Fee,
				Description:   "Transfer fee",
				CreatedAt:     entryTime,
			})
		}
		_ = s.ledger.Append(&domain.LedgerEntry{
			ID:            uuid.NewString(),
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
		s.idempoStore.Put(in.IdempotencyKey, tr.ID)
	}
	return tr, nil
}

func (s *PaymentService) ListHistory(_ context.Context, accountID string, limit int, cursor string) ([]*domain.LedgerEntry, string, error) {
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

func (s *PaymentService) Deposit(_ context.Context, in BalanceChangeInput) (*domain.Account, error) {
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	return s.changeBalance(in.AccountID, in.Currency, in.Amount)
}

func (s *PaymentService) Withdraw(_ context.Context, in BalanceChangeInput) (*domain.Account, error) {
	if in.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	return s.changeBalance(in.AccountID, in.Currency, -in.Amount)
}

func (s *PaymentService) CloseAccount(_ context.Context, accountID string) (*domain.Account, error) {
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

func (s *PaymentService) changeBalance(accountID, currency string, delta domain.Money) (*domain.Account, error) {
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
			ID:            uuid.NewString(),
			AccountID:     accountID,
			OperationType: opType,
			Amount:        amount,
			Description:   desc,
			CreatedAt:     time.Now(),
		}
		_ = s.ledger.Append(entry)
		return acc, nil
	}
	return nil, domain.ErrConflict
}

// ReverseTransfer отменяет выполненный перевод, возвращая средства обратно.
func (s *PaymentService) ReverseTransfer(_ context.Context, transferID string) (*domain.Transfer, error) {
	tr, err := s.transfers.GetTransferByID(transferID)
	if err != nil {
		return nil, err
	}
	if tr.Status != domain.TransferPosted {
		return nil, domain.ErrTransferNotPosted
	}
	if tr.Status == domain.TransferReversed {
		return nil, domain.ErrTransferAlreadyReversed
	}

	// создаем обратный перевод (возвращаем сумму + комиссию)
	reverseTr := &domain.Transfer{
		ID:            uuid.NewString(),
		FromAccountID: tr.ToAccountID,
		ToAccountID:   tr.FromAccountID,
		Amount:        tr.Amount + tr.Fee, // возвращаем сумму + комиссию
		Fee:           0,                  // при реверсе комиссия не взимается
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
		// Отправителю возвращаем сумму + комиссию
		to.Balance += tr.Amount + tr.Fee

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
			ID:            uuid.NewString(),
			AccountID:     from.ID,
			TransferID:    reverseTr.ID,
			OperationType: domain.OperationReversal,
			Amount:        -tr.Amount,
			Description:   "Reversal of transfer " + transferID + " (return amount)",
			CreatedAt:     entryTime,
		})
		// Запись для отправителя: возвращаем сумму перевода
		_ = s.ledger.Append(&domain.LedgerEntry{
			ID:            uuid.NewString(),
			AccountID:     to.ID,
			TransferID:    reverseTr.ID,
			OperationType: domain.OperationReversal,
			Amount:        tr.Amount,
			Description:   "Reversal of transfer " + transferID + " (return amount)",
			CreatedAt:     entryTime,
		})
		// Запись для отправителя: возвращаем комиссию
		if tr.Fee > 0 {
			_ = s.ledger.Append(&domain.LedgerEntry{
				ID:            uuid.NewString(),
				AccountID:     to.ID,
				TransferID:    reverseTr.ID,
				OperationType: domain.OperationReversal,
				Amount:        tr.Fee,
				Description:   "Reversal of transfer " + transferID + " (return fee)",
				CreatedAt:     entryTime,
			})
		}

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
func (s *PaymentService) BatchTransfer(_ context.Context, items []BatchTransferItem, idempotencyKeyPrefix string) (*BatchTransferResult, error) {
	if len(items) == 0 {
		return &BatchTransferResult{}, nil
	}

	result := &BatchTransferResult{
		Total:     len(items),
		Transfers: make([]*domain.Transfer, 0, len(items)),
		Errors:    make([]BatchTransferError, 0),
	}

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

		totalDebit := item.Amount + item.Fee
		if from.Balance < totalDebit {
			result.Errors = append(result.Errors, BatchTransferError{
				Index: i,
				Item:  item,
				Error: domain.ErrInsufficientFunds.Error(),
			})
			continue
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
			idemKey = idempotencyKeyPrefix + ":" + string(rune(i))
		}

		tr, err := s.Transfer(context.Background(), TransferInput{
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

	return result, nil
}
