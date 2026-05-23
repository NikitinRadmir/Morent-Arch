package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"morent-backend/internal/service"
)

// APIError — публичный JSON-ответ об ошибке.
type APIError struct {
	Error  string `json:"error"`
	Code   string `json:"code,omitempty"`
	Status int    `json:"-"`
}

// WriteAPIError пишет JSON с корректным статусом без утечки внутренних деталей.
func WriteAPIError(w http.ResponseWriter, ae APIError) {
	if ae.Error == "" {
		ae.Error = http.StatusText(ae.Status)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.Status)
	_ = json.NewEncoder(w).Encode(ae)
}

func WriteBadRequest(w http.ResponseWriter, message, code string) {
	WriteAPIError(w, APIError{Error: message, Code: code, Status: http.StatusBadRequest})
}

func WriteUnauthorized(w http.ResponseWriter, message string) {
	WriteAPIError(w, APIError{Error: message, Code: "unauthorized", Status: http.StatusUnauthorized})
}

func WriteForbidden(w http.ResponseWriter, message string) {
	WriteAPIError(w, APIError{Error: message, Code: "forbidden", Status: http.StatusForbidden})
}

func WriteNotFound(w http.ResponseWriter, message string) {
	WriteAPIError(w, APIError{Error: message, Code: "not_found", Status: http.StatusNotFound})
}

func WriteValidationError(w http.ResponseWriter, err error) {
	WriteBadRequest(w, ValidationMessage(err), "validation")
}

func ValidationMessage(err error) string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) || len(validationErrors) == 0 {
		return "Проверьте корректность заполнения полей"
	}

	messages := make([]string, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		messages = append(messages, validationFieldMessage(fieldErr))
	}
	return strings.Join(messages, ". ")
}

func validationFieldMessage(err validator.FieldError) string {
	field := err.Field()
	switch field {
	case "Email":
		if err.Tag() == "required" {
			return "Укажите email"
		}
		return "Укажите корректный email"
	case "Password":
		switch err.Tag() {
		case "required":
			return "Укажите пароль"
		case "min":
			return "Пароль слишком короткий"
		case "max":
			return "Пароль слишком длинный"
		}
	case "Name":
		switch err.Tag() {
		case "required":
			return "Укажите имя"
		case "min":
			return "Имя должно быть не короче 2 символов"
		case "max":
			return "Имя слишком длинное"
		}
	case "Code":
		if err.Tag() == "len" {
			return "Код подтверждения должен состоять из 6 символов"
		}
		return "Укажите код подтверждения"
	case "Phone":
		return "Укажите телефон"
	case "CarID":
		return "Выберите автомобиль"
	case "StartDate":
		return "Укажите дату начала аренды"
	case "EndDate":
		return "Укажите дату окончания аренды"
	case "TotalPrice":
		return "Укажите корректную сумму аренды"
	case "Description":
		return "Укажите текст комментария"
	case "Rating":
		return "Укажите оценку от 1 до 5"
	case "Amount":
		return "Укажите корректную сумму"
	case "RecipientCardNumber":
		return "Укажите номер карты получателя"
	}
	return "Проверьте поле " + field
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
	case errors.Is(err, service.ErrGeneratorUnavailable):
		return APIError{Error: "генератор паролей временно недоступен", Code: "generator_unavailable", Status: http.StatusServiceUnavailable}, true
	case errors.Is(err, service.ErrGeneratorBadResponse):
		return APIError{Error: "генератор паролей вернул некорректный ответ", Code: "generator_bad_response", Status: http.StatusBadGateway}, true
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
