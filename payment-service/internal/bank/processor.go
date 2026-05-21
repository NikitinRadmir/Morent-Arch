package bank

import (
	"context"
	"errors"
	"strings"
	"time"

	"morent-arch/payment-service/internal/domain"
	"morent-arch/payment-service/internal/service"
	morentevents "morent-events"

	"github.com/google/uuid"
)

const currencyRUB = "RUB"

type Processor struct {
	pay      *service.PaymentService
	registry *Registry
}

func NewProcessor(pay *service.PaymentService, registry *Registry) *Processor {
	return &Processor{pay: pay, registry: registry}
}

func (p *Processor) Handle(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID, OK: false}
	switch cmd.Type {
	case "bank.register":
		return p.register(ctx, cmd)
	case "bank.login":
		return p.login(ctx, cmd)
	case "bank.logout":
		return p.logout(cmd)
	case "bank.profile":
		return p.profile(ctx, cmd)
	case "bank.deposit":
		return p.deposit(ctx, cmd)
	case "bank.transfer":
		return p.transfer(ctx, cmd)
	case "bank.transactions":
		return p.transactions(ctx, cmd)
	default:
		resp.Error = "unknown command"
		resp.ErrorCode = "unknown_command"
		return resp
	}
}

func (p *Processor) register(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID}
	phone, err := normalizePhone(cmd.Phone)
	if err != nil {
		return fail(resp, err, "invalid_phone")
	}
	if len(strings.TrimSpace(cmd.Password)) < 6 {
		return fail(resp, errors.New("password must be at least 6 characters"), "validation")
	}
	if p.registry.HasPhone(phone) {
		return fail(resp, errors.New("client with this phone already exists"), "phone_exists")
	}
	displayName := strings.TrimSpace(cmd.DisplayName)
	if displayName == "" {
		displayName = "Клиент"
	}
	hash, err := HashPassword(cmd.Password)
	if err != nil {
		return fail(resp, err, "internal")
	}
	acc, err := p.pay.CreateAccount(ctx, service.CreateAccountInput{
		Owner:    phone,
		Currency: currencyRUB,
	})
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	p.registry.RegisterClient(phone, displayName, hash, acc.ID)
	token := uuid.NewString()
	p.registry.CreateSession(token, acc.ID)
	profile, err := p.buildProfile(ctx, acc.ID, phone, displayName)
	if err != nil {
		return fail(resp, err, "internal")
	}
	resp.OK = true
	resp.Token = token
	resp.Profile = profile
	return resp
}

func (p *Processor) login(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID}
	phone, err := normalizePhone(cmd.Phone)
	if err != nil {
		return fail(resp, err, "invalid_phone")
	}
	client, ok := p.registry.VerifyLogin(phone, cmd.Password)
	if !ok {
		return fail(resp, errors.New("invalid phone or password"), "invalid_credentials")
	}
	token := uuid.NewString()
	p.registry.CreateSession(token, client.AccountID)
	profile, err := p.buildProfile(ctx, client.AccountID, client.Phone, client.DisplayName)
	if err != nil {
		return fail(resp, err, "internal")
	}
	resp.OK = true
	resp.Token = token
	resp.Profile = profile
	return resp
}

func (p *Processor) logout(cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID, OK: true}
	if strings.TrimSpace(cmd.Token) != "" {
		p.registry.DeleteSession(cmd.Token)
	}
	return resp
}

func (p *Processor) profile(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID}
	accountID, client, err := p.sessionClient(cmd.Token)
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	profile, err := p.buildProfile(ctx, accountID, client.Phone, client.DisplayName)
	if err != nil {
		return fail(resp, err, "internal")
	}
	resp.OK = true
	resp.Profile = profile
	return resp
}

func (p *Processor) deposit(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID}
	accountID, client, err := p.sessionClient(cmd.Token)
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	amountMinor, err := rubToMinor(cmd.Amount)
	if err != nil {
		return fail(resp, err, "invalid_amount")
	}
	_, err = p.pay.Deposit(ctx, service.BalanceChangeInput{
		AccountID: accountID,
		Amount:    amountMinor,
		Currency:  currencyRUB,
	})
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	profile, err := p.buildProfile(ctx, accountID, client.Phone, client.DisplayName)
	if err != nil {
		return fail(resp, err, "internal")
	}
	resp.OK = true
	resp.Profile = profile
	return resp
}

func (p *Processor) transfer(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID}
	fromID, fromClient, err := p.sessionClient(cmd.Token)
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	recipientPhone, err := normalizePhone(cmd.RecipientPhone)
	if err != nil {
		return fail(resp, err, "invalid_phone")
	}
	if recipientPhone == fromClient.Phone {
		return fail(resp, errors.New("cannot transfer to the same account"), "same_account")
	}
	toClient, ok := p.registry.ClientByPhone(recipientPhone)
	if !ok {
		return fail(resp, errors.New("recipient not found"), "recipient_not_found")
	}
	amountMinor, err := rubToMinor(cmd.Amount)
	if err != nil {
		return fail(resp, err, "invalid_amount")
	}
	_, err = p.pay.Transfer(ctx, service.TransferInput{
		FromAccountID:  fromID,
		ToAccountID:    toClient.AccountID,
		Amount:         amountMinor,
		Fee:            0,
		Currency:       currencyRUB,
		IdempotencyKey: "bank-" + cmd.RequestID,
	})
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	profile, err := p.buildProfile(ctx, fromID, fromClient.Phone, fromClient.DisplayName)
	if err != nil {
		return fail(resp, err, "internal")
	}
	resp.OK = true
	resp.Profile = profile
	return resp
}

func (p *Processor) transactions(ctx context.Context, cmd *morentevents.BankCommand) morentevents.BankResponse {
	resp := morentevents.BankResponse{RequestID: cmd.RequestID}
	accountID, _, err := p.sessionClient(cmd.Token)
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	limit := cmd.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	entries, _, err := p.pay.ListHistory(ctx, accountID, limit, "")
	if err != nil {
		return fail(resp, err, mapDomain(err))
	}
	acc, _ := p.pay.GetAccount(ctx, accountID)
	var balance domain.Money
	if acc != nil {
		balance = acc.Balance
	}
	rows := make([]morentevents.BankTransaction, 0, len(entries))
	for i, e := range entries {
		amount := minorToRub(e.Amount)
		if amount < 0 {
			amount = -amount
		}
		rows = append(rows, morentevents.BankTransaction{
			ID:           uint(i + 1),
			Type:         string(e.OperationType),
			Amount:       amount,
			BalanceAfter: minorToRub(balance),
			CreatedAt:    e.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	resp.OK = true
	resp.Transactions = rows
	return resp
}

func (p *Processor) sessionClient(token string) (string, *clientRecord, error) {
	if strings.TrimSpace(token) == "" {
		return "", nil, errors.New("invalid or expired session")
	}
	accountID, ok := p.registry.AccountByToken(token)
	if !ok {
		return "", nil, errors.New("invalid or expired session")
	}
	client, ok := p.registry.ClientByAccountID(accountID)
	if !ok {
		return "", nil, errors.New("invalid or expired session")
	}
	return accountID, client, nil
}

func (p *Processor) buildProfile(ctx context.Context, accountID, phone, displayName string) (*morentevents.BankProfile, error) {
	acc, err := p.pay.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return &morentevents.BankProfile{
		Phone:       phone,
		DisplayName: displayName,
		Role:        "Клиент",
		Balance:     minorToRub(acc.Balance),
	}, nil
}

func fail(resp morentevents.BankResponse, err error, code string) morentevents.BankResponse {
	resp.OK = false
	resp.Error = err.Error()
	resp.ErrorCode = code
	return resp
}

func mapDomain(err error) string {
	switch {
	case errors.Is(err, domain.ErrInsufficientFunds):
		return "insufficient_funds"
	case errors.Is(err, domain.ErrNotFound):
		return "not_found"
	case errors.Is(err, domain.ErrSameAccount):
		return "same_account"
	case errors.Is(err, domain.ErrInvalidAmount):
		return "invalid_amount"
	case errors.Is(err, domain.ErrConflict):
		return "phone_exists"
	default:
		if strings.Contains(err.Error(), "session") {
			return "invalid_session"
		}
		return "internal"
	}
}

func normalizePhone(raw string) (string, error) {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	d := b.String()
	if len(d) == 0 {
		return "", errors.New("invalid phone")
	}
	if d[0] == '7' {
		d = "8" + d[1:]
	} else if d[0] != '8' {
		d = "8" + d
	}
	if len(d) != 11 {
		return "", errors.New("invalid phone")
	}
	return d, nil
}

func rubToMinor(amount float64) (domain.Money, error) {
	if amount <= 0 {
		return 0, domain.ErrInvalidAmount
	}
	return domain.Money(int64(amount*100 + 0.5)), nil
}

func minorToRub(minor domain.Money) float64 {
	return float64(minor) / 100
}
