package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGeneratorClientDisabled(t *testing.T) {
	client := NewGeneratorClient("")

	if _, err := client.GeneratePassword(context.Background()); !errors.Is(err, ErrGeneratorUnavailable) {
		t.Fatalf("GeneratePassword error = %v, want ErrGeneratorUnavailable", err)
	}
	if _, err := client.ValidatePassword(context.Background(), "Aa1!aaaa"); !errors.Is(err, ErrGeneratorUnavailable) {
		t.Fatalf("ValidatePassword error = %v, want ErrGeneratorUnavailable", err)
	}
}

func TestGeneratorClientBadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewGeneratorClient(server.URL)

	if _, err := client.GeneratePassword(context.Background()); !errors.Is(err, ErrGeneratorBadResponse) {
		t.Fatalf("GeneratePassword error = %v, want ErrGeneratorBadResponse", err)
	}
	if _, err := client.ValidatePassword(context.Background(), "Aa1!aaaa"); !errors.Is(err, ErrGeneratorBadResponse) {
		t.Fatalf("ValidatePassword error = %v, want ErrGeneratorBadResponse", err)
	}
}

func TestGeneratorClientSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/password":
			_, _ = w.Write([]byte(`{"password":"Aa1!aaaa","mask":"********","length":8}`))
		case "/api/v1/password/validate":
			_, _ = w.Write([]byte(`{"valid":true,"hasLower":true,"hasUpper":true,"hasDigit":true,"hasSpecial":true,"lengthOk":true,"length":8}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewGeneratorClient(server.URL)

	password, err := client.GeneratePassword(context.Background())
	if err != nil {
		t.Fatalf("GeneratePassword error = %v", err)
	}
	if password != "Aa1!aaaa" {
		t.Fatalf("password = %q, want Aa1!aaaa", password)
	}

	result, err := client.ValidatePassword(context.Background(), password)
	if err != nil {
		t.Fatalf("ValidatePassword error = %v", err)
	}
	if result == nil || !result.Valid {
		t.Fatalf("ValidatePassword result = %#v, want valid", result)
	}
}
