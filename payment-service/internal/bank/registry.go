package bank

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ClientRecord — данные клиента Morent Bank (карта привязана к счёту в payment-service).
type ClientRecord struct {
	AccountID    string
	Phone        string
	DisplayName  string
	PasswordHash []byte
	CardNumber   string
	ExpDate      string
	CVV          string
	CardHolder   string
}

type sessionRecord struct {
	AccountID string
	Token     string
}

// ClientRegistry — хранилище клиентов и сессий банка.
type ClientRegistry interface {
	HasPhone(phone string) bool
	HasCardNumber(number string) bool
	RegisterClient(phone, displayName, passwordHash, accountID, cardNumber, expDate, cvv, cardHolder string)
	VerifyLogin(phone, password string) (*ClientRecord, bool)
	CreateSession(token, accountID string)
	DeleteSession(token string)
	AccountByToken(token string) (string, bool)
	ClientByPhone(phone string) (*ClientRecord, bool)
	ClientByCardNumber(number string) (*ClientRecord, bool)
	ClientByAccountID(accountID string) (*ClientRecord, bool)
	EnsureClientCard(phone string, registeredAt time.Time) bool
}

// Registry — in-memory реализация (тесты и STORAGE=memory).
type Registry struct {
	mu       sync.RWMutex
	clients  map[string]*ClientRecord
	cards    map[string]string
	sessions map[string]*sessionRecord
}

func NewRegistry() *Registry {
	return &Registry{
		clients:  make(map[string]*ClientRecord),
		cards:    make(map[string]string),
		sessions: make(map[string]*sessionRecord),
	}
}

var _ ClientRegistry = (*Registry)(nil)

func (r *Registry) HasPhone(phone string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.clients[phone]
	return ok
}

func (r *Registry) HasCardNumber(number string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.cards[number]
	return ok
}

func (r *Registry) RegisterClient(phone, displayName, passwordHash, accountID, cardNumber, expDate, cvv, cardHolder string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[phone] = &ClientRecord{
		AccountID:    accountID,
		Phone:        phone,
		DisplayName:  displayName,
		PasswordHash: []byte(passwordHash),
		CardNumber:   cardNumber,
		ExpDate:      expDate,
		CVV:          cvv,
		CardHolder:   cardHolder,
	}
	if cardNumber != "" {
		r.cards[cardNumber] = phone
	}
}

func (r *Registry) VerifyLogin(phone, password string) (*ClientRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clients[phone]
	if !ok {
		return nil, false
	}
	if bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password)) != nil {
		return nil, false
	}
	return c, true
}

func (r *Registry) CreateSession(token, accountID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[token] = &sessionRecord{AccountID: accountID, Token: token}
}

func (r *Registry) DeleteSession(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
}

func (r *Registry) AccountByToken(token string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[token]
	if !ok {
		return "", false
	}
	return s.AccountID, true
}

func (r *Registry) ClientByPhone(phone string) (*ClientRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clients[phone]
	if !ok {
		return nil, false
	}
	return c, true
}

func (r *Registry) ClientByCardNumber(number string) (*ClientRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	phone, ok := r.cards[number]
	if !ok {
		return nil, false
	}
	c, ok := r.clients[phone]
	if !ok {
		return nil, false
	}
	return c, true
}

func (r *Registry) ClientByAccountID(accountID string) (*ClientRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.clients {
		if c.AccountID == accountID {
			return c, true
		}
	}
	return nil, false
}

func (r *Registry) EnsureClientCard(phone string, registeredAt time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clients[phone]
	if !ok || c.CardNumber != "" {
		return ok && c != nil && c.CardNumber != ""
	}
	displayName := c.DisplayName
	number, expDate, cvv, cardHolder, err := GenerateVirtualCard(displayName, registeredAt, func(n string) bool {
		_, exists := r.cards[n]
		return exists
	})
	if err != nil {
		return false
	}
	c.CardNumber = number
	c.ExpDate = expDate
	c.CVV = cvv
	c.CardHolder = cardHolder
	r.cards[number] = phone
	return true
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
