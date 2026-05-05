package service_test

import (
	"context"
	"testing"

	"morent-arch/payment-service/internal/domain"
	"morent-arch/payment-service/internal/repository/memory"
	"morent-arch/payment-service/internal/service"
)

func newTestService() *service.PaymentService {
	store := memory.NewStore()
	return service.NewPaymentService(store, store, store, store, store)
}

func TestDepositAndWithdraw(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	acc, err := svc.CreateAccount(ctx, service.CreateAccountInput{
		Owner:    "Alice",
		Currency: "RUB",
	})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	// депозит 10000
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{
		AccountID: acc.ID,
		Amount:    10000,
		Currency:  "RUB",
	}); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	// списание 4000
	if _, err := svc.Withdraw(ctx, service.BalanceChangeInput{
		AccountID: acc.ID,
		Amount:    4000,
		Currency:  "RUB",
	}); err != nil {
		t.Fatalf("Withdraw: %v", err)
	}

	got, err := svc.GetAccount(ctx, acc.ID)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if got.Balance != 6000 {
		t.Fatalf("balance = %d, want 6000", got.Balance)
	}
}

func TestTransferAndReverse(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	from, err := svc.CreateAccount(ctx, service.CreateAccountInput{Owner: "Alice", Currency: "RUB"})
	if err != nil {
		t.Fatalf("CreateAccount from: %v", err)
	}
	to, err := svc.CreateAccount(ctx, service.CreateAccountInput{Owner: "Bob", Currency: "RUB"})
	if err != nil {
		t.Fatalf("CreateAccount to: %v", err)
	}

	// Депозит отправителю
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{
		AccountID: from.ID,
		Amount:    20000,
		Currency:  "RUB",
	}); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	// Перевод 5000 + комиссия 1000
	tr, err := svc.Transfer(ctx, service.TransferInput{
		FromAccountID: from.ID,
		ToAccountID:   to.ID,
		Amount:        5000,
		Fee:           1000,
		Currency:      "RUB",
	})
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if tr.Status != domain.TransferPosted {
		t.Fatalf("status = %s, want posted", tr.Status)
	}

	fromAfter, _ := svc.GetAccount(ctx, from.ID)
	toAfter, _ := svc.GetAccount(ctx, to.ID)

	if fromAfter.Balance != 14000 { // 20000 - 5000 - 1000
		t.Fatalf("from balance = %d, want 14000", fromAfter.Balance)
	}
	if toAfter.Balance != 5000 {
		t.Fatalf("to balance = %d, want 5000", toAfter.Balance)
	}

	// Реверс
	rev, err := svc.ReverseTransfer(ctx, tr.ID)
	if err != nil {
		t.Fatalf("ReverseTransfer: %v", err)
	}
	if rev.Status != domain.TransferPosted {
		t.Fatalf("reverse transfer status = %s, want posted", rev.Status)
	}

	// Проверяем, что оригинальный перевод помечен как reversed
	trAfter, _ := svc.GetTransfer(ctx, tr.ID)
	if trAfter.Status != domain.TransferReversed {
		t.Fatalf("original transfer status = %s, want reversed", trAfter.Status)
	}

	fromFinal, _ := svc.GetAccount(ctx, from.ID)
	toFinal, _ := svc.GetAccount(ctx, to.ID)

	if fromFinal.Balance != 20000 {
		t.Fatalf("from final balance = %d, want 20000", fromFinal.Balance)
	}
	if toFinal.Balance != 0 {
		t.Fatalf("to final balance = %d, want 0", toFinal.Balance)
	}
}

func TestGetAccountSummary(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	acc, err := svc.CreateAccount(ctx, service.CreateAccountInput{Owner: "Alice", Currency: "RUB"})
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	// два депозита и одно списание
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{AccountID: acc.ID, Amount: 1000, Currency: "RUB"}); err != nil {
		t.Fatalf("Deposit1: %v", err)
	}
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{AccountID: acc.ID, Amount: 2000, Currency: "RUB"}); err != nil {
		t.Fatalf("Deposit2: %v", err)
	}
	if _, err := svc.Withdraw(ctx, service.BalanceChangeInput{AccountID: acc.ID, Amount: 500, Currency: "RUB"}); err != nil {
		t.Fatalf("Withdraw: %v", err)
	}

	summary, err := svc.GetAccountSummary(ctx, acc.ID)
	if err != nil {
		t.Fatalf("GetAccountSummary: %v", err)
	}

	if summary.TotalDeposits != 3000 {
		t.Fatalf("TotalDeposits = %d, want 3000", summary.TotalDeposits)
	}
	if summary.TotalWithdraws != 500 {
		t.Fatalf("TotalWithdraws = %d, want 500", summary.TotalWithdraws)
	}
}
