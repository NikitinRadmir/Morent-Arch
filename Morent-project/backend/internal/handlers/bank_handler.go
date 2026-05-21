package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"morent-backend/internal/config"
	bankdto "morent-backend/internal/modules/bank/httpdto"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/service"
)

type BankHandler struct {
	bank *service.BankService
	auth *service.AuthService
	cfg  *config.Config
}

func NewBankHandler(bank *service.BankService, auth *service.AuthService, cfg *config.Config) *BankHandler {
	return &BankHandler{bank: bank, auth: auth, cfg: cfg}
}

var bankValidator = validator.New()

func (h *BankHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req bankdto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := bankValidator.Struct(req); err != nil {
		http.Error(w, "validation error", http.StatusBadRequest)
		return
	}
	profile, token, err := h.bank.Register(req.Phone, req.Password, req.DisplayName)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.setBankCookie(w, token)
	h.writeJSON(w, http.StatusCreated, bankdto.AuthResponse{Token: token, Profile: *profile})
}

func (h *BankHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req bankdto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := bankValidator.Struct(req); err != nil {
		http.Error(w, "validation error", http.StatusBadRequest)
		return
	}
	profile, token, err := h.bank.Login(req.Phone, req.Password)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.setBankCookie(w, token)
	h.writeJSON(w, http.StatusOK, bankdto.AuthResponse{Token: token, Profile: *profile})
}

func (h *BankHandler) SyncSession(w http.ResponseWriter, r *http.Request) {
	user, err := common.Authenticate(h.auth, h.cfg, r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	displayName := strings.TrimSpace(user.Nickname)
	if displayName == "" {
		displayName = strings.TrimSpace(user.Name)
	}
	profile, token, err := h.bank.EnsureSessionForUser(user.ID, displayName)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.setBankCookie(w, token)
	h.writeJSON(w, http.StatusOK, profile)
}

func (h *BankHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := common.BankSessionToken(r, h.cfg)
	_ = h.bank.Logout(token)
	h.clearBankCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *BankHandler) Profile(w http.ResponseWriter, r *http.Request) {
	token := common.BankSessionToken(r, h.cfg)
	profile, err := h.bank.Profile(token)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, profile)
}

func (h *BankHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	token := common.BankSessionToken(r, h.cfg)
	var req bankdto.AmountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	profile, err := h.bank.Deposit(token, req.Amount)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, bankdto.OperationResponse{
		Message: "deposit completed",
		Profile: *profile,
	})
}

func (h *BankHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	token := common.BankSessionToken(r, h.cfg)
	var req bankdto.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	profile, err := h.bank.Transfer(token, req.RecipientCardNumber, req.Amount)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, bankdto.OperationResponse{
		Message: "transfer completed",
		Profile: *profile,
	})
}

func (h *BankHandler) Transactions(w http.ResponseWriter, r *http.Request) {
	token := common.BankSessionToken(r, h.cfg)
	limit := 20
	if q := r.URL.Query().Get("limit"); q != "" {
		if n, err := strconv.Atoi(q); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	rows, err := h.bank.ListTransactions(token, limit)
	if err != nil {
		h.writeBankError(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, rows)
}

func (h *BankHandler) setBankCookie(w http.ResponseWriter, token string) {
	if token == "" {
		return
	}
	name := h.cfg.BankSessionCookieName
	if name == "" {
		name = "morent_bank_session"
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.SessionCookieSecure,
		Domain:   h.cfg.SessionCookieDomain,
	})
}

func (h *BankHandler) clearBankCookie(w http.ResponseWriter) {
	name := h.cfg.BankSessionCookieName
	if name == "" {
		name = "morent_bank_session"
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Domain:   h.cfg.SessionCookieDomain,
	})
}

func (h *BankHandler) writeBankError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrBankUnavailable):
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	case errors.Is(err, service.ErrBankPhoneExists):
		http.Error(w, err.Error(), http.StatusConflict)
		return
	case errors.Is(err, service.ErrBankInvalidCredentials), errors.Is(err, service.ErrBankSessionInvalid):
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	case errors.Is(err, service.ErrBankInvalidPhone), errors.Is(err, service.ErrBankInvalidCard),
		errors.Is(err, service.ErrBankInvalidAmount),
		errors.Is(err, service.ErrBankInsufficientFunds), errors.Is(err, service.ErrBankRecipientNotFound),
		errors.Is(err, service.ErrBankSameAccount):
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	default:
		if strings.Contains(err.Error(), "password must be") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		common.WriteInternalError(w, "bank operation failed")
	}
}

func (h *BankHandler) writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
