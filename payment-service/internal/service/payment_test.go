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
	return service.NewPaymentService(store, store, store, store, store, store)
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

	if fromFinal.Balance != 19000 {
		t.Fatalf("from final balance = %d, want 19000", fromFinal.Balance)
	}
	if toFinal.Balance != 0 {
		t.Fatalf("to final balance = %d, want 0", toFinal.Balance)
	}
}

func TestTransferIdempotency(t *testing.T) {
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
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{AccountID: from.ID, Amount: 10000, Currency: "RUB"}); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	first, err := svc.Transfer(ctx, service.TransferInput{
		FromAccountID:  from.ID,
		ToAccountID:    to.ID,
		Amount:         2500,
		Currency:       "RUB",
		IdempotencyKey: "rent-42",
	})
	if err != nil {
		t.Fatalf("first transfer: %v", err)
	}
	second, err := svc.Transfer(ctx, service.TransferInput{
		FromAccountID:  from.ID,
		ToAccountID:    to.ID,
		Amount:         2500,
		Currency:       "RUB",
		IdempotencyKey: "rent-42",
	})
	if err != nil {
		t.Fatalf("second transfer: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("idempotent transfer ID = %s, want %s", second.ID, first.ID)
	}

	fromAfter, _ := svc.GetAccount(ctx, from.ID)
	toAfter, _ := svc.GetAccount(ctx, to.ID)
	if fromAfter.Balance != 7500 {
		t.Fatalf("from balance = %d, want 7500", fromAfter.Balance)
	}
	if toAfter.Balance != 2500 {
		t.Fatalf("to balance = %d, want 2500", toAfter.Balance)
	}
}

func TestTransferIdempotencyRejectsDifferentPayload(t *testing.T) {
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
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{AccountID: from.ID, Amount: 10000, Currency: "RUB"}); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	if _, err := svc.Transfer(ctx, service.TransferInput{
		FromAccountID:  from.ID,
		ToAccountID:    to.ID,
		Amount:         2500,
		Currency:       "RUB",
		IdempotencyKey: "rent-43",
	}); err != nil {
		t.Fatalf("first transfer: %v", err)
	}
	if _, err := svc.Transfer(ctx, service.TransferInput{
		FromAccountID:  from.ID,
		ToAccountID:    to.ID,
		Amount:         3000,
		Currency:       "RUB",
		IdempotencyKey: "rent-43",
	}); err != domain.ErrIdempotencyKeyConflict {
		t.Fatalf("second transfer error = %v, want idempotency key conflict", err)
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
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

	if _, err := svc.Transfer(ctx, service.TransferInput{FromAccountID: from.ID, ToAccountID: to.ID, Amount: 1, Currency: "RUB"}); err != domain.ErrInsufficientFunds {
		t.Fatalf("Transfer error = %v, want insufficient funds", err)
	}
}

func TestCreatePaymentIdempotency(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	first, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
		ReferenceID:    "rental-42",
		UserID:         "7",
		CarID:          "11",
		Amount:         125000,
		Currency:       "RUB",
		IdempotencyKey: "payment-rental-42",
	})
	if err != nil {
		t.Fatalf("CreatePayment first: %v", err)
	}
	second, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
		ReferenceID:    "rental-42",
		UserID:         "7",
		CarID:          "11",
		Amount:         125000,
		Currency:       "RUB",
		IdempotencyKey: "payment-rental-42",
	})
	if err != nil {
		t.Fatalf("CreatePayment second: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("idempotent payment ID = %s, want %s", second.ID, first.ID)
	}
	if second.Status != domain.PaymentSucceeded {
		t.Fatalf("payment status = %s, want succeeded", second.Status)
	}

	got, err := svc.GetPayment(ctx, first.ID)
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if got.ReferenceID != "rental-42" {
		t.Fatalf("reference ID = %s, want rental-42", got.ReferenceID)
	}
}

func TestCreatePaymentIdempotencyRejectsDifferentPayload(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	if _, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
		ReferenceID:    "rental-43",
		UserID:         "7",
		CarID:          "11",
		Amount:         125000,
		Currency:       "RUB",
		IdempotencyKey: "payment-rental-43",
	}); err != nil {
		t.Fatalf("CreatePayment first: %v", err)
	}
	if _, err := svc.CreatePayment(ctx, service.CreatePaymentInput{
		ReferenceID:    "rental-43",
		UserID:         "7",
		CarID:          "11",
		Amount:         126000,
		Currency:       "RUB",
		IdempotencyKey: "payment-rental-43",
	}); err != domain.ErrIdempotencyKeyConflict {
		t.Fatalf("CreatePayment second error = %v, want idempotency key conflict", err)
	}
}

func TestBatchTransferRejectsAggregateOverdraft(t *testing.T) {
	svc := newTestService()
	ctx := context.Background()

	from, err := svc.CreateAccount(ctx, service.CreateAccountInput{Owner: "Alice", Currency: "RUB"})
	if err != nil {
		t.Fatalf("CreateAccount from: %v", err)
	}
	to1, err := svc.CreateAccount(ctx, service.CreateAccountInput{Owner: "Bob", Currency: "RUB"})
	if err != nil {
		t.Fatalf("CreateAccount to1: %v", err)
	}
	to2, err := svc.CreateAccount(ctx, service.CreateAccountInput{Owner: "Carol", Currency: "RUB"})
	if err != nil {
		t.Fatalf("CreateAccount to2: %v", err)
	}
	if _, err := svc.Deposit(ctx, service.BalanceChangeInput{AccountID: from.ID, Amount: 10000, Currency: "RUB"}); err != nil {
		t.Fatalf("Deposit: %v", err)
	}

	result, err := svc.BatchTransfer(ctx, []service.BatchTransferItem{
		{FromAccountID: from.ID, ToAccountID: to1.ID, Amount: 7000, Currency: "RUB"},
		{FromAccountID: from.ID, ToAccountID: to2.ID, Amount: 7000, Currency: "RUB"},
	}, "batch-1")
	if err != nil {
		t.Fatalf("BatchTransfer: %v", err)
	}
	if result.Failed == 0 {
		t.Fatalf("Failed = 0, want aggregate overdraft failure")
	}
	fromAfter, _ := svc.GetAccount(ctx, from.ID)
	if fromAfter.Balance != 10000 {
		t.Fatalf("from balance = %d, want 10000", fromAfter.Balance)
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
