package messaging

import (
	"errors"

	morentevents "morent-events"
	"morent-backend/internal/bank"
)

// BankResponse — ответ payment-service по Kafka.
type BankResponse struct {
	RequestID    string
	OK           bool
	Error        string
	ErrorCode    string
	Token        string
	Profile      *bank.ProfileResponse
	Transactions []bank.TransactionResponse
}

func MapBankResponse(raw *morentevents.BankResponse) BankResponse {
	if raw == nil {
		return BankResponse{OK: false, Error: "empty response"}
	}
	out := BankResponse{
		RequestID: raw.RequestID,
		OK:        raw.OK,
		Error:     raw.Error,
		ErrorCode: raw.ErrorCode,
		Token:     raw.Token,
	}
	if raw.Profile != nil {
		out.Profile = &bank.ProfileResponse{
			Phone:       raw.Profile.Phone,
			DisplayName: raw.Profile.DisplayName,
			Role:        raw.Profile.Role,
			Balance:     raw.Profile.Balance,
		}
	}
	if len(raw.Transactions) > 0 {
		out.Transactions = make([]bank.TransactionResponse, len(raw.Transactions))
		for i, t := range raw.Transactions {
			out.Transactions[i] = bank.TransactionResponse{
				ID:                t.ID,
				Type:              t.Type,
				Amount:            t.Amount,
				CounterpartyPhone: t.CounterpartyPhone,
				BalanceAfter:      t.BalanceAfter,
				CreatedAt:         t.CreatedAt,
			}
		}
	}
	return out
}

func (r BankResponse) AsError() error {
	if r.OK {
		return nil
	}
	msg := r.Error
	if msg == "" {
		msg = "bank operation failed"
	}
	switch r.ErrorCode {
	case "phone_exists":
		return ErrBankPhoneExists{msg}
	case "invalid_credentials":
		return ErrBankInvalidCredentials{msg}
	case "invalid_session":
		return ErrBankSessionInvalid{msg}
	case "invalid_phone":
		return ErrBankInvalidPhone{msg}
	case "invalid_amount":
		return ErrBankInvalidAmount{msg}
	case "insufficient_funds":
		return ErrBankInsufficientFunds{msg}
	case "recipient_not_found":
		return ErrBankRecipientNotFound{msg}
	case "same_account":
		return ErrBankSameAccount{msg}
	default:
		return errors.New(msg)
	}
}

type ErrBankPhoneExists struct{ Msg string }

func (e ErrBankPhoneExists) Error() string { return e.Msg }

type ErrBankInvalidCredentials struct{ Msg string }

func (e ErrBankInvalidCredentials) Error() string { return e.Msg }

type ErrBankSessionInvalid struct{ Msg string }

func (e ErrBankSessionInvalid) Error() string { return e.Msg }

type ErrBankInvalidPhone struct{ Msg string }

func (e ErrBankInvalidPhone) Error() string { return e.Msg }

type ErrBankInvalidAmount struct{ Msg string }

func (e ErrBankInvalidAmount) Error() string { return e.Msg }

type ErrBankInsufficientFunds struct{ Msg string }

func (e ErrBankInsufficientFunds) Error() string { return e.Msg }

type ErrBankRecipientNotFound struct{ Msg string }

func (e ErrBankRecipientNotFound) Error() string { return e.Msg }

type ErrBankSameAccount struct{ Msg string }

func (e ErrBankSameAccount) Error() string { return e.Msg }
