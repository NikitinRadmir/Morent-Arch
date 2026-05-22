package postgres

import (
	"errors"
	"sort"
	"strings"
	"time"

	"morent-arch/payment-service/internal/domain"
	"morent-arch/payment-service/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

var (
	_ repository.AccountRepository     = (*Store)(nil)
	_ repository.TransferRepository    = (*Store)(nil)
	_ repository.LedgerRepository      = (*Store)(nil)
	_ repository.PaymentRepository     = (*Store)(nil)
	_ repository.IdempotencyRepository = (*Store)(nil)
	_ repository.LimitsRepository      = (*Store)(nil)
)

func (s *Store) CreateAccount(a *domain.Account) error {
	if a.Status == "" {
		a.Status = domain.AccountActive
	}
	now := time.Now()
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}
	a.UpdatedAt = now
	if a.Version == 0 {
		a.Version = 1
	}
	row := accountToRow(a)
	err := s.db.Create(&row).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (s *Store) GetAccountByID(id string) (*domain.Account, error) {
	var row AccountRow
	if err := s.db.First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToAccount(&row), nil
}

func (s *Store) UpdateAccount(a *domain.Account) error {
	expected := a.Version
	a.Version++
	a.UpdatedAt = time.Now()
	row := accountToRow(a)
	res := s.db.Model(&AccountRow{}).
		Where("id = ? AND version = ?", a.ID, expected).
		Updates(map[string]any{
			"owner":         row.Owner,
			"currency":      row.Currency,
			"balance":       row.Balance,
			"status":        row.Status,
			"version":       row.Version,
			"daily_limit":   row.DailyLimit,
			"monthly_limit": row.MonthlyLimit,
			"updated_at":    row.UpdatedAt,
			"closed_at":     row.ClosedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) ListAccounts() ([]*domain.Account, error) {
	var rows []AccountRow
	if err := s.db.Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Account, len(rows))
	for i := range rows {
		out[i] = rowToAccount(&rows[i])
	}
	return out, nil
}

func (s *Store) CreateTransfer(t *domain.Transfer) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	row := transferToRow(t)
	err := s.db.Create(&row).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (s *Store) GetTransferByID(id string) (*domain.Transfer, error) {
	var row TransferRow
	if err := s.db.First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToTransfer(&row), nil
}

func (s *Store) UpdateTransfer(t *domain.Transfer) error {
	row := transferToRow(t)
	res := s.db.Model(&TransferRow{}).Where("id = ?", t.ID).Updates(map[string]any{
		"status":      row.Status,
		"posted_at":   row.PostedAt,
		"reversed_at": row.ReversedAt,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) ListTransfers(filter repository.TransferFilter) ([]*domain.Transfer, error) {
	q := s.db.Model(&TransferRow{})
	if filter.FromAccountID != "" {
		q = q.Where("from_account_id = ?", filter.FromAccountID)
	}
	if filter.ToAccountID != "" {
		q = q.Where("to_account_id = ?", filter.ToAccountID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.FromDate != nil {
		q = q.Where("created_at >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		q = q.Where("created_at <= ?", *filter.ToDate)
	}
	var rows []TransferRow
	if err := q.Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*domain.Transfer, len(rows))
	for i := range rows {
		out[i] = rowToTransfer(&rows[i])
	}
	return out, nil
}

func (s *Store) CreatePayment(p *domain.Payment) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	row := paymentToRow(p)
	err := s.db.Create(&row).Error
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrConflict
		}
		return err
	}
	return nil
}

func (s *Store) GetPaymentByID(id string) (*domain.Payment, error) {
	var row PaymentRow
	if err := s.db.First(&row, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToPayment(&row), nil
}

func (s *Store) Append(e *domain.LedgerEntry) error {
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	row := LedgerEntryRow{
		ID:            e.ID,
		AccountID:     e.AccountID,
		TransferID:    e.TransferID,
		OperationType: string(e.OperationType),
		Amount:        e.Amount,
		Description:   e.Description,
		CreatedAt:     e.CreatedAt,
	}
	return s.db.Create(&row).Error
}

func (s *Store) ListByAccount(accountID string, limit int, cursor string) ([]*domain.LedgerEntry, string, error) {
	q := s.db.Where("account_id = ?", accountID).Order("created_at asc")
	var rows []LedgerEntryRow
	if err := q.Find(&rows).Error; err != nil {
		return nil, "", err
	}
	entries := make([]*domain.LedgerEntry, len(rows))
	for i := range rows {
		entries[i] = rowToLedger(&rows[i])
	}
	start := 0
	if cursor != "" {
		for i, e := range entries {
			if e.ID == cursor {
				start = i + 1
				break
			}
		}
	}
	end := start + limit
	if end > len(entries) {
		end = len(entries)
	}
	var next string
	if end < len(entries) && end > 0 {
		next = entries[end-1].ID
	}
	out := make([]*domain.LedgerEntry, end-start)
	copy(out, entries[start:end])
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, next, nil
}

func (s *Store) Get(key string) (repository.IdempotencyRecord, bool) {
	var row IdempotencyRow
	if err := s.db.First(&row, "key = ?", key).Error; err != nil {
		return repository.IdempotencyRecord{}, false
	}
	return repository.IdempotencyRecord{
		Key:          row.Key,
		ResourceID:   row.ResourceID,
		ResourceType: row.ResourceType,
		Fingerprint:  row.Fingerprint,
	}, true
}

func (s *Store) Put(record repository.IdempotencyRecord) {
	row := IdempotencyRow{
		Key:          record.Key,
		ResourceID:   record.ResourceID,
		ResourceType: record.ResourceType,
		Fingerprint:  record.Fingerprint,
	}
	_ = s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"resource_id", "resource_type", "fingerprint"}),
	}).Create(&row).Error
}

func (s *Store) GetOperationLimits(accountID string, period domain.LimitPeriod, date time.Time) (*domain.OperationLimits, error) {
	key, periodStart, periodEnd := limitKey(accountID, period, date)
	var row OperationLimitRow
	if err := s.db.First(&row, "key = ?", key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &domain.OperationLimits{
				AccountID:   accountID,
				Period:      period,
				SpentAmount: 0,
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			}, nil
		}
		return nil, err
	}
	return &domain.OperationLimits{
		AccountID:   row.AccountID,
		Period:      domain.LimitPeriod(row.Period),
		SpentAmount: domain.Money(row.SpentAmount),
		PeriodStart: row.PeriodStart,
		PeriodEnd:   row.PeriodEnd,
	}, nil
}

func (s *Store) RecordOperation(accountID string, period domain.LimitPeriod, amount domain.Money, date time.Time) error {
	key, periodStart, periodEnd := limitKey(accountID, period, date)
	var row OperationLimitRow
	err := s.db.First(&row, "key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.db.Create(&OperationLimitRow{
			Key: key, AccountID: accountID, Period: string(period),
			SpentAmount: int64(amount), PeriodStart: periodStart, PeriodEnd: periodEnd,
		}).Error
	}
	if err != nil {
		return err
	}
	return s.db.Model(&row).Update("spent_amount", row.SpentAmount+int64(amount)).Error
}

func limitKey(accountID string, period domain.LimitPeriod, date time.Time) (string, time.Time, time.Time) {
	if period == domain.LimitDaily {
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		return accountID + ":daily:" + start.Format("2006-01-02"), start, start.Add(24 * time.Hour)
	}
	start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	return accountID + ":monthly:" + start.Format("2006-01"), start, start.AddDate(0, 1, 0)
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return errors.Is(err, gorm.ErrDuplicatedKey) ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "UNIQUE constraint")
}

func accountToRow(a *domain.Account) AccountRow {
	return AccountRow{
		ID: a.ID, Owner: a.Owner, Currency: a.Currency, Balance: int64(a.Balance),
		Status: string(a.Status), Version: a.Version, DailyLimit: int64(a.DailyLimit),
		MonthlyLimit: int64(a.MonthlyLimit), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt, ClosedAt: a.ClosedAt,
	}
}

func rowToAccount(r *AccountRow) *domain.Account {
	return &domain.Account{
		ID: r.ID, Owner: r.Owner, Currency: r.Currency, Balance: domain.Money(r.Balance),
		Status: domain.AccountStatus(r.Status), Version: r.Version,
		DailyLimit: domain.Money(r.DailyLimit), MonthlyLimit: domain.Money(r.MonthlyLimit),
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, ClosedAt: r.ClosedAt,
	}
}

func transferToRow(t *domain.Transfer) TransferRow {
	return TransferRow{
		ID: t.ID, FromAccountID: t.FromAccountID, ToAccountID: t.ToAccountID,
		FromCardNumber: t.FromCardNumber, ToCardNumber: t.ToCardNumber,
		Amount: int64(t.Amount), Fee: int64(t.Fee), Currency: t.Currency, Status: string(t.Status),
		CreatedAt: t.CreatedAt, PostedAt: t.PostedAt, ReversedAt: t.ReversedAt,
	}
}

func rowToTransfer(r *TransferRow) *domain.Transfer {
	return &domain.Transfer{
		ID: r.ID, FromAccountID: r.FromAccountID, ToAccountID: r.ToAccountID,
		FromCardNumber: r.FromCardNumber, ToCardNumber: r.ToCardNumber,
		Amount: domain.Money(r.Amount), Fee: domain.Money(r.Fee), Currency: r.Currency,
		Status: domain.TransferStatus(r.Status), CreatedAt: r.CreatedAt,
		PostedAt: r.PostedAt, ReversedAt: r.ReversedAt,
	}
}

func paymentToRow(p *domain.Payment) PaymentRow {
	return PaymentRow{
		ID: p.ID, ReferenceID: p.ReferenceID, AccountID: p.AccountID, CardNumber: p.CardNumber,
		UserID: p.UserID, CarID: p.CarID, Amount: int64(p.Amount), Currency: p.Currency,
		Status: string(p.Status), CreatedAt: p.CreatedAt, ProcessedAt: p.ProcessedAt,
	}
}

func rowToPayment(r *PaymentRow) *domain.Payment {
	return &domain.Payment{
		ID: r.ID, ReferenceID: r.ReferenceID, AccountID: r.AccountID, CardNumber: r.CardNumber,
		UserID: r.UserID, CarID: r.CarID, Amount: domain.Money(r.Amount), Currency: r.Currency,
		Status: domain.PaymentStatus(r.Status), CreatedAt: r.CreatedAt, ProcessedAt: r.ProcessedAt,
	}
}

func rowToLedger(r *LedgerEntryRow) *domain.LedgerEntry {
	return &domain.LedgerEntry{
		ID: r.ID, AccountID: r.AccountID, TransferID: r.TransferID,
		OperationType: domain.OperationType(r.OperationType), Amount: domain.Money(r.Amount),
		Description: r.Description, CreatedAt: r.CreatedAt,
	}
}
