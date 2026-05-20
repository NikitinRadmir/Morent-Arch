package common

import (
	"net/http"
	"strings"

	"morent-backend/internal/config"
	"morent-backend/internal/models"
	"morent-backend/internal/service"
)

// SessionToken извлекает токен сессии из Authorization (Bearer) или HttpOnly cookie.
func SessionToken(r *http.Request, cfg *config.Config) string {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		return strings.TrimSpace(token[7:])
	}
	if token != "" {
		return token
	}
	cookieName := cfg.SessionCookieName
	if cookieName == "" {
		cookieName = "morent_session"
	}
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie == nil {
		return ""
	}
	return strings.TrimSpace(cookie.Value)
}

// Authenticate возвращает пользователя по сессионному токену.
func Authenticate(auth *service.AuthService, cfg *config.Config, r *http.Request) (*models.User, error) {
	token := SessionToken(r, cfg)
	if token == "" {
		return nil, service.ErrInvalidToken
	}
	user, err := auth.GetUserByToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, service.ErrInvalidToken
	}
	return user, nil
}

// WithAuth требует валидную сессию (cookie или Bearer).
func WithAuth(auth *service.AuthService, cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := Authenticate(auth, cfg, r); err != nil {
			status := http.StatusUnauthorized
			msg := "unauthorized"
			if err == service.ErrInvalidToken {
				msg = "invalid or missing session"
			}
			http.Error(w, msg, status)
			return
		}
		next(w, r)
	}
}
