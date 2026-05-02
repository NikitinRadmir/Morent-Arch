package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"morent-arch/payment-service/internal/domain"
	"morent-arch/payment-service/internal/repository"
	"morent-arch/payment-service/internal/repository/memory"
	"morent-arch/payment-service/internal/service"
)

type Handler struct {
	pay *service.PaymentService
}

func NewRouter() http.Handler {
	store := memory.NewStore()
	h := &Handler{pay: service.NewPaymentService(store, store, store, store, store)}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/v1/accounts", h.accounts)
	mux.HandleFunc("/api/v1/accounts/", h.accountByID)
	mux.HandleFunc("/api/v1/owners/", h.ownerBalance)
	mux.HandleFunc("/api/v1/transfers", h.transfers)
	mux.HandleFunc("/api/v1/transfers/", h.transferByID)

	return withMiddleware(mux)
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) accounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createAccount(w, r)
	case http.MethodGet:
		h.listAccounts(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) accountByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts/")
	parts := splitPath(path)
	if len(parts) == 0 || parts[0] == "" {
		notFound(w)
		return
	}
	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.getAccount(w, r, id)
		case http.MethodDelete:
			h.closeAccount(w, r, id)
		default:
			methodNotAllowed(w)
		}
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "summary":
			if r.Method == http.MethodGet {
				h.getAccountSummary(w, r, id)
				return
			}
		case "history":
			if r.Method == http.MethodGet {
				h.getHistory(w, r, id)
				return
			}
		case "deposit":
			if r.Method == http.MethodPost {
				h.deposit(w, r, id)
				return
			}
		case "withdraw":
			if r.Method == http.MethodPost {
				h.withdraw(w, r, id)
				return
			}
		}
	}
	notFound(w)
}

func (h *Handler) ownerBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/owners/")
	parts := splitPath(path)
	if len(parts) != 2 || parts[1] != "balance" {
		notFound(w)
		return
	}
	owner, err := urlPathUnescape(parts[0])
	if err != nil || strings.TrimSpace(owner) == "" {
		badRequest(w, "invalid owner")
		return
	}
	balance, err := h.pay.GetOwnerBalance(r.Context(), owner)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ownerBalanceResponse{Owner: balance.Owner, Balances: balance.Balances})
}

func (h *Handler) transfers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/transfers" {
		notFound(w)
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.createTransfer(w, r)
	case http.MethodGet:
		h.listTransfers(w, r)
	default:
		methodNotAllowed(w)
	}
}

func (h *Handler) transferByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/transfers/")
	parts := splitPath(path)
	if len(parts) == 1 && parts[0] == "stats" && r.Method == http.MethodGet {
		h.transferStats(w, r)
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		h.getTransfer(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[1] == "reverse" && r.Method == http.MethodPost {
		h.reverseTransfer(w, r, parts[0])
		return
	}
	notFound(w)
}

func (h *Handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	acc, err := h.pay.CreateAccount(r.Context(), service.CreateAccountInput{Owner: req.Owner, Currency: req.Currency, DailyLimit: req.DailyLimit, MonthlyLimit: req.MonthlyLimit})
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAccountResponse(acc))
}

func (h *Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	filter := service.ListAccountsFilter{Owner: r.URL.Query().Get("owner")}
	if status := r.URL.Query().Get("status"); status != "" {
		if status != string(domain.AccountActive) && status != string(domain.AccountClosed) {
			badRequest(w, "unknown status filter")
			return
		}
		filter.Status = domain.AccountStatus(status)
	}
	items, err := h.pay.ListAccounts(r.Context(), filter)
	if err != nil {
		domainError(w, err)
		return
	}
	resp := make([]accountResponse, 0, len(items))
	for _, acc := range items {
		resp = append(resp, toAccountResponse(acc))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getAccount(w http.ResponseWriter, r *http.Request, id string) {
	acc, err := h.pay.GetAccount(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAccountResponse(acc))
}

func (h *Handler) closeAccount(w http.ResponseWriter, r *http.Request, id string) {
	acc, err := h.pay.CloseAccount(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAccountResponse(acc))
}

func (h *Handler) getAccountSummary(w http.ResponseWriter, r *http.Request, id string) {
	s, err := h.pay.GetAccountSummary(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, accountSummaryResponse{Account: toAccountResponse(s.Account), TotalDeposits: s.TotalDeposits, TotalWithdraws: s.TotalWithdraws, TotalInTransfers: s.TotalInTransfers, TotalOutTransfers: s.TotalOutTransfers, TotalFees: s.TotalFees})
}

func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request, id string) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			badRequest(w, "invalid limit")
			return
		}
		limit = n
	}
	items, next, err := h.pay.ListHistory(r.Context(), id, limit, r.URL.Query().Get("cursor"))
	if err != nil {
		domainError(w, err)
		return
	}
	resp := ledgerResponse{Entries: make([]ledgerEntryResponse, 0, len(items)), NextCursor: next}
	for _, e := range items {
		resp.Entries = append(resp.Entries, toLedgerEntryResponse(e))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) deposit(w http.ResponseWriter, r *http.Request, id string) {
	h.changeBalance(w, r, id, true)
}

func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request, id string) {
	h.changeBalance(w, r, id, false)
}

func (h *Handler) changeBalance(w http.ResponseWriter, r *http.Request, id string, deposit bool) {
	var req balanceChangeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	var (
		acc *domain.Account
		err error
	)
	input := service.BalanceChangeInput{AccountID: id, Amount: req.Amount, Currency: req.Currency}
	if deposit {
		acc, err = h.pay.Deposit(r.Context(), input)
	} else {
		acc, err = h.pay.Withdraw(r.Context(), input)
	}
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toAccountResponse(acc))
}

func (h *Handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	var req transferRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	tr, err := h.pay.Transfer(r.Context(), service.TransferInput{FromAccountID: req.FromAccountID, ToAccountID: req.ToAccountID, Amount: req.Amount, Fee: req.Fee, Currency: req.Currency, IdempotencyKey: r.Header.Get("Idempotency-Key")})
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTransferResponse(tr))
}

func (h *Handler) listTransfers(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseTransferFilter(w, r)
	if !ok {
		return
	}
	items, err := h.pay.ListTransfers(r.Context(), filter)
	if err != nil {
		domainError(w, err)
		return
	}
	resp := make([]transferResponse, 0, len(items))
	for _, tr := range items {
		resp = append(resp, toTransferResponse(tr))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getTransfer(w http.ResponseWriter, r *http.Request, id string) {
	tr, err := h.pay.GetTransfer(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTransferResponse(tr))
}

func (h *Handler) reverseTransfer(w http.ResponseWriter, r *http.Request, id string) {
	tr, err := h.pay.ReverseTransfer(r.Context(), id)
	if err != nil {
		domainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTransferResponse(tr))
}

func (h *Handler) transferStats(w http.ResponseWriter, r *http.Request) {
	filter, ok := parseTransferFilter(w, r)
	if !ok {
		return
	}
	stats, err := h.pay.GetTransferStats(r.Context(), filter)
	if err != nil {
		domainError(w, err)
		return
	}
	resp := transferStatsResponse{TotalTransfers: stats.TotalTransfers, TotalAmount: stats.TotalAmount, PostedTransfers: stats.PostedTransfers, ReversedTransfers: stats.ReversedTransfers}
	if stats.FromDate != nil {
		resp.FromDate = stats.FromDate.Format(time.RFC3339)
	}
	if stats.ToDate != nil {
		resp.ToDate = stats.ToDate.Format(time.RFC3339)
	}
	writeJSON(w, http.StatusOK, resp)
}

func parseTransferFilter(w http.ResponseWriter, r *http.Request) (repository.TransferFilter, bool) {
	q := r.URL.Query()
	filter := repository.TransferFilter{FromAccountID: q.Get("from_account_id"), ToAccountID: q.Get("to_account_id")}
	if status := q.Get("status"); status != "" {
		s := domain.TransferStatus(status)
		if s != domain.TransferPending && s != domain.TransferPosted && s != domain.TransferReversed {
			badRequest(w, "unknown status filter")
			return filter, false
		}
		filter.Status = s
	}
	if raw := q.Get("from_date"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			badRequest(w, "invalid from_date; use RFC3339")
			return filter, false
		}
		filter.FromDate = &t
	}
	if raw := q.Get("to_date"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			badRequest(w, "invalid to_date; use RFC3339")
			return filter, false
		}
		filter.ToDate = &t
	}
	return filter, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		badRequest(w, "invalid json: "+err.Error())
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func domainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrConflict):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInsufficientFunds), errors.Is(err, domain.ErrCurrencyMismatch), errors.Is(err, domain.ErrSameAccount), errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrAccountClosed), errors.Is(err, domain.ErrUnsupportedCurrency), errors.Is(err, domain.ErrInvalidOwner), errors.Is(err, domain.ErrTransferNotPosted), errors.Is(err, domain.ErrTransferAlreadyReversed), errors.Is(err, domain.ErrLimitExceeded), errors.Is(err, domain.ErrInvalidFee):
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal"})
	}
}

func badRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: message})
}
func notFound(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
}
func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
}
func splitPath(path string) []string           { return strings.Split(strings.Trim(path, "/"), "/") }
func urlPathUnescape(s string) (string, error) { return strings.ReplaceAll(s, "%20", " "), nil }
