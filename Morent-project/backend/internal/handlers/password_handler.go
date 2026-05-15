package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"morent-backend/internal/service"
)

// PasswordHandler проксирует генерацию/валидацию пароля к generator-service.
// Фронт ходит только сюда (тот же origin), пароль при валидации — только в POST body.
type PasswordHandler struct {
	generator *service.GeneratorClient
	log       *slog.Logger
}

func NewPasswordHandler(generator *service.GeneratorClient, log *slog.Logger) *PasswordHandler {
	if log == nil {
		log = slog.Default()
	}
	return &PasswordHandler{generator: generator, log: log}
}

func (h *PasswordHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.generator.Enabled() {
		http.Error(w, "generator service unavailable", http.StatusServiceUnavailable)
		return
	}
	password, err := h.generator.GeneratePassword(r.Context())
	if err != nil {
		h.log.Warn("password generate failed", "error", err)
		http.Error(w, "generator service unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]string{"password": password})
}

type validatePasswordBody struct {
	Password string `json:"password"`
}

func (h *PasswordHandler) Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 512))
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var body validatePasswordBody
	if err := json.Unmarshal(raw, &body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	if !h.generator.Enabled() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"unavailable": true,
			"message":     "generator service unavailable",
		})
		return
	}

	result, err := h.generator.ValidatePassword(r.Context(), body.Password)
	if err != nil {
		h.log.Warn("password validate failed", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"unavailable": true,
			"message":     "generator service unavailable",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(result)
}
