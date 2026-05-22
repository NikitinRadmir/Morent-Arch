package common

import (
	"encoding/json"
	"errors"
	"net/http"

	"morent-backend/internal/service"
)

// APIError — публичный JSON-ответ об ошибке.
type APIError struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Status  int    `json:"-"`
}

// WriteAPIError пишет JSON с корректным статусом без утечки внутренних деталей.
func WriteAPIError(w http.ResponseWriter, ae APIError) {
	if ae.Error == "" {
		ae.Error = http.StatusText(ae.Status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.Status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": ae.Error,
		"code":  ae.Code,
	})
}

// MapServiceError сопоставляет известные доменные ошибки HTTP-кодам.
func MapServiceError(err error) (APIError, bool) {
	if err == nil {
		return APIError{}, false
	}
	switch {
	case errors.Is(err, service.ErrUserExists):
		return APIError{Error: "пользователь уже существует", Code: "user_exists", Status: http.StatusConflict}, true
	case errors.Is(err, service.ErrInvalidCredentials):
		return APIError{Error: "неверный email или пароль", Code: "invalid_credentials", Status: http.StatusUnauthorized}, true
	case errors.Is(err, service.ErrEmailNotVerified):
		return APIError{Error: "подтвердите email", Code: "email_not_verified", Status: http.StatusForbidden}, true
	case errors.Is(err, service.ErrBankUnavailable):
		return APIError{Error: "банковский сервис временно недоступен", Code: "bank_unavailable", Status: http.StatusServiceUnavailable}, true
	case errors.Is(err, service.ErrBankSessionRequired), errors.Is(err, service.ErrBankSessionInvalid):
		return APIError{Error: "требуется вход в Morent Bank", Code: "bank_session_required", Status: http.StatusUnauthorized}, true
	case errors.Is(err, service.ErrInsufficientBankBalance), errors.Is(err, service.ErrBankInsufficientFunds):
		return APIError{Error: "недостаточно средств на счёте", Code: "insufficient_funds", Status: http.StatusPaymentRequired}, true
	default:
		return APIError{}, false
	}
}

// RespondError — typed mapping или безопасный 500.
func RespondError(w http.ResponseWriter, err error) {
	if ae, ok := MapServiceError(err); ok {
		WriteAPIError(w, ae)
		return
	}
	WriteInternalErrorJSON(w, "внутренняя ошибка сервера")
}

// WriteInternalErrorJSON — 500 без текста err.Error().
func WriteInternalErrorJSON(w http.ResponseWriter, publicMessage string) {
	WriteAPIError(w, APIError{
		Error:  publicMessage,
		Code:   "internal",
		Status: http.StatusInternalServerError,
	})
}
