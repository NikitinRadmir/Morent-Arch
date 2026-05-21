package morentevents

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	TopicBankCommands  = "morent.bank.commands"
	TopicBankResponses = "morent.bank.responses"

	SourceMorentBackend   = "morent-backend"
	SourcePaymentService  = "payment-service"
)

// BankCommand — запрос Morent → payment-service (топик morent.bank.commands).
type BankCommand struct {
	Type           string    `json:"type"`
	RequestID      string    `json:"request_id"`
	Phone          string    `json:"phone,omitempty"`
	Password       string    `json:"password,omitempty"`
	DisplayName    string    `json:"display_name,omitempty"`
	Token          string    `json:"token,omitempty"`
	Amount         float64   `json:"amount,omitempty"`
	RecipientPhone string    `json:"recipient_phone,omitempty"`
	Limit          int       `json:"limit,omitempty"`
	SentAt         time.Time `json:"sent_at,omitempty"`
}

// BankProfile — профиль клиента в ответе.
type BankProfile struct {
	Phone       string  `json:"phone"`
	DisplayName string  `json:"displayName"`
	Role        string  `json:"role"`
	Balance     float64 `json:"balance"`
}

// BankTransaction — операция в истории.
type BankTransaction struct {
	ID                uint    `json:"id"`
	Type              string  `json:"type"`
	Amount            float64 `json:"amount"`
	CounterpartyPhone string  `json:"counterpartyPhone,omitempty"`
	BalanceAfter      float64 `json:"balanceAfter"`
	CreatedAt         string  `json:"createdAt"`
}

// BankResponse — ответ payment-service → Morent (топик morent.bank.responses).
type BankResponse struct {
	RequestID    string            `json:"request_id"`
	OK           bool              `json:"ok"`
	Error        string            `json:"error,omitempty"`
	ErrorCode    string            `json:"error_code,omitempty"`
	Token        string            `json:"token,omitempty"`
	Profile      *BankProfile      `json:"profile,omitempty"`
	Transactions []BankTransaction `json:"transactions,omitempty"`
}

func MarshalBankCommand(cmd BankCommand) ([]byte, error) {
	out, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal bank command: %w", err)
	}
	return out, nil
}

func UnmarshalBankCommand(raw []byte) (*BankCommand, error) {
	var cmd BankCommand
	if err := json.Unmarshal(raw, &cmd); err != nil {
		return nil, fmt.Errorf("unmarshal bank command: %w", err)
	}
	if cmd.RequestID == "" || cmd.Type == "" {
		return nil, fmt.Errorf("invalid bank command")
	}
	return &cmd, nil
}

func MarshalBankResponse(resp BankResponse) ([]byte, error) {
	out, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal bank response: %w", err)
	}
	return out, nil
}

func UnmarshalBankResponse(raw []byte) (*BankResponse, error) {
	var resp BankResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal bank response: %w", err)
	}
	if resp.RequestID == "" {
		return nil, fmt.Errorf("invalid bank response: missing request_id")
	}
	return &resp, nil
}
