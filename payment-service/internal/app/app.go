package app

import (
	"fmt"
	"log/slog"

	"morent-arch/payment-service/internal/bank"
	"morent-arch/payment-service/internal/config"
	"morent-arch/payment-service/internal/repository/memory"
	"morent-arch/payment-service/internal/repository/postgres"
	"morent-arch/payment-service/internal/service"
	"morent-arch/payment-service/internal/storage"

	"gorm.io/gorm"
)

// Deps — общие зависимости HTTP API и Kafka consumer банка.
type Deps struct {
	Payment   *service.PaymentService
	Processor *bank.Processor
	DB        *gorm.DB
}

func Bootstrap(cfg config.Config) (*Deps, error) {
	if cfg.Storage == "memory" {
		store := memory.NewStore()
		reg := bank.NewRegistry()
		pay := service.NewPaymentService(store, store, store, store, store, store)
		return &Deps{
			Payment:   pay,
			Processor: bank.NewProcessor(pay, reg),
		}, nil
	}

	db, err := storage.OpenPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}
	store := postgres.NewStore(db)
	reg := postgres.NewBankRepository(db)
	pay := service.NewPaymentService(store, store, store, store, store, store)
	slog.Info("payment-service storage", "driver", "postgres")
	return &Deps{
		Payment:   pay,
		Processor: bank.NewProcessor(pay, reg),
		DB:        db,
	}, nil
}
