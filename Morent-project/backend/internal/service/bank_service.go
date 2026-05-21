package service

import (
	"context"
	"errors"
	"strings"

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

func (s *BankService) publish(ctx context.Context, cmd messaging.BankCommand) error {
	if err := s.unavailable(); err != nil {
		return err
	}
	return s.gateway.Publish(ctx, cmd)
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
	if _, err := NormalizeBankPhone(phone); err != nil {
		return nil, "", err
	}
	if len(strings.TrimSpace(password)) < 6 {
		return nil, "", errors.New("password must be at least 6 characters")
	}
	return nil, "", s.publish(context.Background(), messaging.BankCommand{
		Type:        messaging.BankCmdRegister,
		RequestID:   uuid.NewString(),
		Phone:       phone,
		Password:    password,
		DisplayName: displayName,
	})
}

func (s *BankService) Login(phone, password string) (*bank.ProfileResponse, string, error) {
	if _, err := NormalizeBankPhone(phone); err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(password) == "" {
		return nil, "", ErrBankInvalidCredentials
	}
	return nil, "", s.publish(context.Background(), messaging.BankCommand{
		Type:      messaging.BankCmdLogin,
		RequestID: uuid.NewString(),
		Phone:     phone,
		Password:  password,
	})
}

func (s *BankService) Logout(token string) error {
	if strings.TrimSpace(token) == "" {
		return ErrBankSessionInvalid
	}
	return s.publish(context.Background(), messaging.BankCommand{
		Type:      messaging.BankCmdLogout,
		RequestID: uuid.NewString(),
		Token:     token,
	})
}

func (s *BankService) Profile(token string) (*bank.ProfileResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if err := s.publish(context.Background(), messaging.BankCommand{
		Type:      messaging.BankCmdGetProfile,
		RequestID: uuid.NewString(),
		Token:     token,
	}); err != nil {
		return nil, err
	}
	return nil, ErrBankUnavailable
}

func (s *BankService) Deposit(token string, amount float64) (*bank.ProfileResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if amount <= 0 {
		return nil, ErrBankInvalidAmount
	}
	return nil, s.publish(context.Background(), messaging.BankCommand{
		Type:      messaging.BankCmdDeposit,
		RequestID: uuid.NewString(),
		Token:     token,
		Amount:    amount,
	})
}

func (s *BankService) Transfer(token, recipientPhone string, amount float64) (*bank.ProfileResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if amount <= 0 {
		return nil, ErrBankInvalidAmount
	}
	if _, err := NormalizeBankPhone(recipientPhone); err != nil {
		return nil, err
	}
	return nil, s.publish(context.Background(), messaging.BankCommand{
		Type:           messaging.BankCmdTransfer,
		RequestID:      uuid.NewString(),
		Token:          token,
		Amount:         amount,
		RecipientPhone: recipientPhone,
	})
}

func (s *BankService) ListTransactions(token string, limit int) ([]bank.TransactionResponse, error) {
	if strings.TrimSpace(token) == "" {
		return nil, ErrBankSessionInvalid
	}
	if limit <= 0 {
		limit = 20
	}
	if err := s.publish(context.Background(), messaging.BankCommand{
		Type:      messaging.BankCmdTransactions,
		RequestID: uuid.NewString(),
		Token:     token,
		Limit:     limit,
	}); err != nil {
		return nil, err
	}
	return nil, ErrBankUnavailable
}
