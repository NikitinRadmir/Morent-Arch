package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"morent-backend/internal/bank"
	"morent-backend/internal/messaging"
)

var (
	ErrBankUnavailable        = errors.New("bank is not available until Kafka integration is enabled")
	ErrBankInvalidPhone       = errors.New("invalid phone number")
	ErrBankInvalidAmount      = errors.New("invalid amount")
	ErrBankSessionInvalid     = errors.New("invalid or expired session")
	ErrBankPhoneExists        = errors.New("client with this phone already exists")
	ErrBankInvalidCredentials = errors.New("invalid phone or password")
	ErrBankInsufficientFunds  = errors.New("insufficient funds")
	ErrBankRecipientNotFound  = errors.New("recipient not found")
	ErrBankSameAccount        = errors.New("cannot transfer to the same account")
)

type BankService struct {
	gateway messaging.BankGateway
}

func NewBankService(gateway messaging.BankGateway) *BankService {
	if gateway == nil {
		gateway = messaging.BankNoopGateway{}
	}
	return &BankService{gateway: gateway}
}

func (s *BankService) Available() bool {
	return s.gateway.Enabled()
}

func (s *BankService) unavailable() error {
	if s.Available() {
		return nil
	}
	return ErrBankUnavailable
}

func (s *BankService) request(ctx context.Context, cmd messaging.BankCommand) (messaging.BankResponse, error) {
	if err := s.unavailable(); err != nil {
		return messaging.BankResponse{}, err
	}
	if cmd.RequestID == "" {
		cmd.RequestID = uuid.NewString()
	}
	if cmd.SentAt.IsZero() {
		cmd.SentAt = time.Now().UTC()
	}
	resp, err := s.gateway.Request(ctx, cmd)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return messaging.BankResponse{}, ErrBankUnavailable
		}
		return messaging.BankResponse{}, err
	}
	if err := mapBankResponseError(resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func mapBankResponseError(resp messaging.BankResponse) error {
	if resp.OK {
		return nil
	}
	err := resp.AsError()
	return mapBankGatewayError(err)
}

func mapBankGatewayError(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case messaging.ErrBankPhoneExists:
		return ErrBankPhoneExists
	case messaging.ErrBankInvalidCredentials:
		return ErrBankInvalidCredentials
	case messaging.ErrBankSessionInvalid:
		return ErrBankSessionInvalid
	case messaging.ErrBankInvalidPhone:
		return ErrBankInvalidPhone
	case messaging.ErrBankInvalidAmount:
		return ErrBankInvalidAmount
	case messaging.ErrBankInsufficientFunds:
		return ErrBankInsufficientFunds
	case messaging.ErrBankRecipientNotFound:
		return ErrBankRecipientNotFound
	case messaging.ErrBankSameAccount:
		return ErrBankSameAccount
	default:
		return err
	}
}

func NormalizeBankPhone(raw string) (string, error) {
	d := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, raw)
	if len(d) == 0 {
		return "", ErrBankInvalidPhone
	}
	if d[0] == '7' {
		d = "8" + d[1:]
	} else if d[0] != '8' {
		d = "8" + d
	}
	if len(d) != 11 {
		return "", ErrBankInvalidPhone
	}
	return d, nil
}

func (s *BankService) Register(phone, password, displayName string) (*bank.ProfileResponse, string, error) {
	normalized, err := NormalizeBankPhone(phone)
	if err != nil {
		return nil, "", err
	}
	if len(strings.TrimSpace(password)) < 6 {
		return nil, "", errors.New("password must be at least 6 characters")
	}
	resp, err := s.request(context.Background(), messaging.BankCommand{
		Type:        messaging.BankCmdRegister,
		Phone:       normalized,
		Password:    password,
		DisplayName: displayName,
	})
	if err != nil {
		return nil, "", err
	}
	return resp.Profile, resp.Token, nil
}

func (s *BankService) Login(phone, password string) (*bank.ProfileResponse, string, error) {
	normalized, err := NormalizeBankPhone(phone)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(password) == "" {
		return nil, "", ErrBankInvalidCredentials
	}
	resp, err := s.request(context.Background(), messaging.BankCommand{
		Type:     messaging.BankCmdLogin,
		Phone:    normalized,
		Password: password,
	})
	if err != nil {
		return nil, "", err
	}
	return resp.Profile, resp.Token, nil
}

func (s *BankService) Logout(token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrBankSessionInvalid
	}
	_, err := s.request(context.Background(), messaging.BankCommand{
		Type:  messaging.BankCmdLogout,
		Token: token,
	})
	return err
}

func (s *BankService) Profile(token string) (*bank.ProfileResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	resp, err := s.request(context.Background(), messaging.BankCommand{
		Type:  messaging.BankCmdGetProfile,
		Token: token,
	})
	if err != nil {
		return nil, err
	}
	return resp.Profile, nil
}

func (s *BankService) Deposit(token string, amount float64) (*bank.ProfileResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if amount <= 0 {
		return nil, ErrBankInvalidAmount
	}
	resp, err := s.request(context.Background(), messaging.BankCommand{
		Type:   messaging.BankCmdDeposit,
		Token:  token,
		Amount: amount,
	})
	if err != nil {
		return nil, err
	}
	return resp.Profile, nil
}

func (s *BankService) Transfer(token, recipientPhone string, amount float64) (*bank.ProfileResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if amount <= 0 {
		return nil, ErrBankInvalidAmount
	}
	normalized, err := NormalizeBankPhone(recipientPhone)
	if err != nil {
		return nil, err
	}
	resp, err := s.request(context.Background(), messaging.BankCommand{
		Type:           messaging.BankCmdTransfer,
		Token:          token,
		Amount:         amount,
		RecipientPhone: normalized,
	})
	if err != nil {
		return nil, err
	}
	return resp.Profile, nil
}

func (s *BankService) ListTransactions(token string, limit int) ([]bank.TransactionResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if limit <= 0 {
		limit = 20
	}
	resp, err := s.request(context.Background(), messaging.BankCommand{
		Type:  messaging.BankCmdTransactions,
		Token: token,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	return resp.Transactions, nil
}
