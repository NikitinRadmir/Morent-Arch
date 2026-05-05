package memory

import (
	"sort"
	"sync"
	"time"

	"morent-arch/payment-service/internal/domain"
	"morent-arch/payment-service/internal/repository"
)

type Store struct {
	mu            sync.RWMutex
	accounts      map[string]*domain.Account
	transfers     map[string]*domain.Transfer
	ledger        map[string][]*domain.LedgerEntry // by account
	idempotency   map[string]string                // key -> transferID
	accountIndex  []string
	transferIndex []string
	// Лимиты: ключ = accountID:period:date (например, "acc123:daily:2025-01-15")
	limits map[string]*domain.OperationLimits
}

func NewStore() *Store {
	return &Store{
		accounts:      make(map[string]*domain.Account),
		transfers:     make(map[string]*domain.Transfer),
		ledger:        make(map[string][]*domain.LedgerEntry),
		idempotency:   make(map[string]string),
		transferIndex: make([]string, 0),
		limits:        make(map[string]*domain.OperationLimits),
	}
}

// Interface guards
var _ repository.AccountRepository = (*Store)(nil)
var _ repository.TransferRepository = (*Store)(nil)
var _ repository.LedgerRepository = (*Store)(nil)
var _ repository.IdempotencyRepository = (*Store)(nil)
var _ repository.LimitsRepository = (*Store)(nil)

func (s *Store) CreateAccount(a *domain.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.accounts[a.ID]; ok {
		return domain.ErrConflict
	}
	if a.Status == "" {
		a.Status = domain.AccountActive
	}
	a.CreatedAt = time.Now()
	a.UpdatedAt = a.CreatedAt
	a.Version = 1
	s.accounts[a.ID] = cloneAccount(a)
	s.accountIndex = append(s.accountIndex, a.ID)
	return nil
}

func (s *Store) GetAccountByID(id string) (*domain.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.accounts[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneAccount(a), nil
}

func (s *Store) UpdateAccount(a *domain.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stored, ok := s.accounts[a.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if stored.Version != a.Version {
		return domain.ErrConflict
	}
	a.Version++
	a.UpdatedAt = time.Now()
	s.accounts[a.ID] = cloneAccount(a)
	return nil
}

func (s *Store) ListAccounts() ([]*domain.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]*domain.Account, 0, len(s.accounts))
	for _, id := range s.accountIndex {
		items = append(items, cloneAccount(s.accounts[id]))
	}
	return items, nil
}

func (s *Store) CreateTransfer(t *domain.Transfer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.transfers[t.ID]; ok {
		return domain.ErrConflict
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	s.transfers[t.ID] = cloneTransfer(t)
	s.transferIndex = append(s.transferIndex, t.ID)
	return nil
}

func (s *Store) GetTransferByID(id string) (*domain.Transfer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.transfers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneTransfer(t), nil
}

func (s *Store) UpdateTransfer(t *domain.Transfer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.transfers[t.ID]; !ok {
		return domain.ErrNotFound
	}
	s.transfers[t.ID] = cloneTransfer(t)
	return nil
}

func (s *Store) ListTransfers(filter repository.TransferFilter) ([]*domain.Transfer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*domain.Transfer
	for _, id := range s.transferIndex {
		t := s.transfers[id]
		if filter.FromAccountID != "" && t.FromAccountID != filter.FromAccountID {
			continue
		}
		if filter.ToAccountID != "" && t.ToAccountID != filter.ToAccountID {
			continue
		}
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.FromDate != nil && t.CreatedAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && t.CreatedAt.After(*filter.ToDate) {
			continue
		}
		out = append(out, cloneTransfer(t))
	}
	// сортируем по дате создания (новые первыми)
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (s *Store) Append(e *domain.LedgerEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.CreatedAt = time.Now()
	s.ledger[e.AccountID] = append(s.ledger[e.AccountID], cloneEntry(e))
	return nil
}

func (s *Store) ListByAccount(accountID string, limit int, cursor string) ([]*domain.LedgerEntry, string, error) {
	s.mu.RLock()
	rows := make([]*domain.LedgerEntry, 0, len(s.ledger[accountID]))
	for _, e := range s.ledger[accountID] {
		rows = append(rows, cloneEntry(e))
	}
	s.mu.RUnlock()

	sort.Slice(rows, func(i, j int) bool { return rows[i].CreatedAt.Before(rows[j].CreatedAt) })

	start := 0
	if cursor != "" {
		for i, e := range rows {
			if e.ID == cursor {
				start = i + 1
				break
			}
		}
	}
	end := start + limit
	if end > len(rows) {
		end = len(rows)
	}
	var next string
	if end < len(rows) && end > 0 {
		next = rows[end-1].ID
	}
	out := make([]*domain.LedgerEntry, end-start)
	copy(out, rows[start:end])
	return out, next, nil
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.idempotency[key]
	return v, ok
}

func (s *Store) Put(key, transferID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idempotency[key] = transferID
}

func cloneAccount(a *domain.Account) *domain.Account       { cp := *a; return &cp }
func cloneTransfer(t *domain.Transfer) *domain.Transfer    { cp := *t; return &cp }
func cloneEntry(e *domain.LedgerEntry) *domain.LedgerEntry { cp := *e; return &cp }

func (s *Store) GetOperationLimits(accountID string, period domain.LimitPeriod, date time.Time) (*domain.OperationLimits, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var key string
	var periodStart, periodEnd time.Time

	if period == domain.LimitDaily {
		periodStart = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		periodEnd = periodStart.Add(24 * time.Hour)
		key = accountID + ":daily:" + periodStart.Format("2006-01-02")
	} else {
		periodStart = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		periodEnd = periodStart.AddDate(0, 1, 0)
		key = accountID + ":monthly:" + periodStart.Format("2006-01")
	}

	lim, ok := s.limits[key]
	if !ok {
		return &domain.OperationLimits{
			AccountID:   accountID,
			Period:      period,
			SpentAmount: 0,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
		}, nil
	}
	return &domain.OperationLimits{
		AccountID:   lim.AccountID,
		Period:      lim.Period,
		SpentAmount: lim.SpentAmount,
		PeriodStart: lim.PeriodStart,
		PeriodEnd:   lim.PeriodEnd,
	}, nil
}

func (s *Store) RecordOperation(accountID string, period domain.LimitPeriod, amount domain.Money, date time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var key string
	var periodStart, periodEnd time.Time

	if period == domain.LimitDaily {
		periodStart = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		periodEnd = periodStart.Add(24 * time.Hour)
		key = accountID + ":daily:" + periodStart.Format("2006-01-02")
	} else {
		periodStart = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
		periodEnd = periodStart.AddDate(0, 1, 0)
		key = accountID + ":monthly:" + periodStart.Format("2006-01")
	}

	lim, ok := s.limits[key]
	if !ok {
		s.limits[key] = &domain.OperationLimits{
			AccountID:   accountID,
			Period:      period,
			SpentAmount: amount,
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
		}
	} else {
		lim.SpentAmount += amount
	}
	return nil
}
