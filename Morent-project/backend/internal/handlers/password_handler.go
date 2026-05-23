package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"morent-backend/internal/modules/transport/http/common"
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
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}
	if !h.generator.Enabled() {
		common.RespondError(w, service.ErrGeneratorUnavailable)
		return
	}
	password, err := h.generator.GeneratePassword(r.Context())
	if err != nil {
		h.log.Warn("password generate failed", "error", err)
		common.RespondError(w, err)
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
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 512))
	if err != nil {
		common.WriteAPIError(w, common.APIError{Error: "invalid request body", Code: "invalid_body", Status: http.StatusBadRequest})
		return
	}
	var body validatePasswordBody
	if err := json.Unmarshal(raw, &body); err != nil {
		common.WriteAPIError(w, common.APIError{Error: "invalid JSON", Code: "invalid_json", Status: http.StatusBadRequest})
		return
	}
	if body.Password == "" {
		common.WriteAPIError(w, common.APIError{Error: "password is required", Code: "password_required", Status: http.StatusBadRequest})
		return
	}

	if !h.generator.Enabled() {
		common.RespondError(w, service.ErrGeneratorUnavailable)
		return
	}

	result, err := h.generator.ValidatePassword(r.Context(), body.Password)
	if err != nil {
		h.log.Warn("password validate failed", "error", err)
		common.RespondError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(result)
}
