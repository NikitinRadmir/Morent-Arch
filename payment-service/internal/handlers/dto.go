package handlers

import (
	"time"

	"morent-arch/payment-service/internal/domain"
)

type createAccountRequest struct {
	Owner        string       `json:"owner"`
	Currency     string       `json:"currency"`
	DailyLimit   domain.Money `json:"daily_limit_minor,omitempty"`
	MonthlyLimit domain.Money `json:"monthly_limit_minor,omitempty"`
}

type balanceChangeRequest struct {
	Amount   domain.Money `json:"amount_minor"`
	Currency string       `json:"currency"`
}

type transferRequest struct {
	FromAccountID string       `json:"from_account_id"`
	ToAccountID   string       `json:"to_account_id"`
	Amount        domain.Money `json:"amount_minor"`
	Fee           domain.Money `json:"fee_minor,omitempty"`
	Currency      string       `json:"currency"`
}

type accountResponse struct {
	ID           string       `json:"id"`
	Owner        string       `json:"owner"`
	Currency     string       `json:"currency"`
	Balance      domain.Money `json:"balance_minor"`
	Version      int64        `json:"version"`
	Status       string       `json:"status"`
	DailyLimit   domain.Money `json:"daily_limit_minor,omitempty"`
	MonthlyLimit domain.Money `json:"monthly_limit_minor,omitempty"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
	ClosedAt     string       `json:"closed_at,omitempty"`
}

type transferResponse struct {
	ID            string       `json:"id"`
	FromAccountID string       `json:"from_account_id"`
	ToAccountID   string       `json:"to_account_id"`
	Amount        domain.Money `json:"amount_minor"`
	Fee           domain.Money `json:"fee_minor"`
	Currency      string       `json:"currency"`
	Status        string       `json:"status"`
	CreatedAt     string       `json:"created_at"`
	PostedAt      string       `json:"posted_at,omitempty"`
	ReversedAt    string       `json:"reversed_at,omitempty"`
}

type accountSummaryResponse struct {
	Account           accountResponse `json:"account"`
	TotalDeposits     domain.Money    `json:"total_deposits_minor"`
	TotalWithdraws    domain.Money    `json:"total_withdraws_minor"`
	TotalInTransfers  domain.Money    `json:"total_in_transfers_minor"`
	TotalOutTransfers domain.Money    `json:"total_out_transfers_minor"`
	TotalFees         domain.Money    `json:"total_fees_minor"`
}

type ledgerEntryResponse struct {
	ID            string       `json:"id"`
	AccountID     string       `json:"account_id"`
	TransferID    string       `json:"transfer_id,omitempty"`
	OperationType string       `json:"operation_type"`
	Amount        domain.Money `json:"amount_minor"`
	Description   string       `json:"description"`
	CreatedAt     string       `json:"created_at"`
}

type ledgerResponse struct {
	Entries    []ledgerEntryResponse `json:"entries"`
	NextCursor string                `json:"next_cursor,omitempty"`
}

type ownerBalanceResponse struct {
	Owner    string                  `json:"owner"`
	Balances map[string]domain.Money `json:"balances_minor"`
}

type transferStatsResponse struct {
	TotalTransfers    int          `json:"total_transfers"`
	TotalAmount       domain.Money `json:"total_amount_minor"`
	PostedTransfers   int          `json:"posted_transfers"`
	ReversedTransfers int          `json:"reversed_transfers"`
	FromDate          string       `json:"from_date,omitempty"`
	ToDate            string       `json:"to_date,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toAccountResponse(acc *domain.Account) accountResponse {
	resp := accountResponse{ID: acc.ID, Owner: acc.Owner, Currency: acc.Currency, Balance: acc.Balance, Version: acc.Version, Status: string(acc.Status), CreatedAt: acc.CreatedAt.Format(time.RFC3339Nano), UpdatedAt: acc.UpdatedAt.Format(time.RFC3339Nano), DailyLimit: acc.DailyLimit, MonthlyLimit: acc.MonthlyLimit}
	if acc.ClosedAt != nil {
		resp.ClosedAt = acc.ClosedAt.Format(time.RFC3339Nano)
	}
	return resp
}

func toTransferResponse(tr *domain.Transfer) transferResponse {
	resp := transferResponse{ID: tr.ID, FromAccountID: tr.FromAccountID, ToAccountID: tr.ToAccountID, Amount: tr.Amount, Fee: tr.Fee, Currency: tr.Currency, Status: string(tr.Status), CreatedAt: tr.CreatedAt.Format(time.RFC3339Nano)}
	if tr.PostedAt != nil {
		resp.PostedAt = tr.PostedAt.Format(time.RFC3339Nano)
	}
	if tr.ReversedAt != nil {
		resp.ReversedAt = tr.ReversedAt.Format(time.RFC3339Nano)
	}
	return resp
}

func toLedgerEntryResponse(e *domain.LedgerEntry) ledgerEntryResponse {
	return ledgerEntryResponse{ID: e.ID, AccountID: e.AccountID, TransferID: e.TransferID, OperationType: string(e.OperationType), Amount: e.Amount, Description: e.Description, CreatedAt: e.CreatedAt.Format(time.RFC3339Nano)}
}
