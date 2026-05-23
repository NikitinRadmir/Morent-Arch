package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"morent-backend/internal/service"
)

func TestMapServiceErrorGeneratorUnavailable(t *testing.T) {
	apiErr, ok := MapServiceError(service.ErrGeneratorUnavailable)
	if !ok {
		t.Fatal("MapServiceError returned ok=false")
	}
	if apiErr.Status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", apiErr.Status, http.StatusServiceUnavailable)
	}
	if apiErr.Code != "generator_unavailable" {
		t.Fatalf("code = %q, want generator_unavailable", apiErr.Code)
	}
}

func TestMapServiceErrorGeneratorBadResponse(t *testing.T) {
	apiErr, ok := MapServiceError(service.ErrGeneratorBadResponse)
	if !ok {
		t.Fatal("MapServiceError returned ok=false")
	}
	if apiErr.Status != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", apiErr.Status, http.StatusBadGateway)
	}
	if apiErr.Code != "generator_bad_response" {
		t.Fatalf("code = %q, want generator_bad_response", apiErr.Code)
	}
}

func TestWriteAPIErrorOmitsEmptyCode(t *testing.T) {
	rec := httptest.NewRecorder()

	WriteAPIError(rec, APIError{Error: "bad request", Status: http.StatusBadRequest})

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode failed: %v", err)
	}
	if _, ok := body["code"]; ok {
		t.Fatalf("code should be omitted for empty code, got %#v", body)
	}
}

func TestValidationMessageIsPublic(t *testing.T) {
	type loginRequest struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required,min=6"`
	}

	err := validator.New().Struct(loginRequest{Email: "user@example.com", Password: "123"})
	message := ValidationMessage(err)

	if message != "Пароль слишком короткий" {
		t.Fatalf("message = %q, want public password message", message)
	}
}
