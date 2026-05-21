package common

import (
	"net/http"
	"strings"

	"morent-backend/internal/config"
	"morent-backend/internal/service"
)

func BankSessionToken(r *http.Request, cfg *config.Config) string {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		return strings.TrimSpace(token[7:])
	}
	if token != "" {
		return token
	}
	name := cfg.BankSessionCookieName
	if name == "" {
		name = "morent_bank_session"
	}
	cookie, err := r.Cookie(name)
	if err != nil || cookie == nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

func WithBankAuth(bank *service.BankService, cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !bank.Available() {
			http.Error(w, service.ErrBankUnavailable.Error(), http.StatusServiceUnavailable)
			return
		}
		if BankSessionToken(r, cfg) == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
