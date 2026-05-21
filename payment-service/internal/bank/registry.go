package bank

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type clientRecord struct {
	AccountID    string
	Phone        string
	DisplayName  string
	PasswordHash []byte
}

type sessionRecord struct {
	AccountID string
	Token     string
}

type Registry struct {
	mu       sync.RWMutex
	clients  map[string]*clientRecord  // phone -> client
	sessions map[string]*sessionRecord // token -> session
}

func NewRegistry() *Registry {
	return &Registry{
		clients:  make(map[string]*clientRecord),
		sessions: make(map[string]*sessionRecord),
	}
}

func (r *Registry) HasPhone(phone string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.clients[phone]
	return ok
}

func (r *Registry) RegisterClient(phone, displayName, passwordHash, accountID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[phone] = &clientRecord{
		AccountID:    accountID,
		Phone:        phone,
		DisplayName:  displayName,
		PasswordHash: []byte(passwordHash),
	}
}

func (r *Registry) VerifyLogin(phone, password string) (*clientRecord, bool) {
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

func (r *Registry) ClientByPhone(phone string) (*clientRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.clients[phone]
	if !ok {
		return nil, false
	}
	return c, true
}

func (r *Registry) ClientByAccountID(accountID string) (*clientRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, c := range r.clients {
		if c.AccountID == accountID {
			return c, true
		}
	}
	return nil, false
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
