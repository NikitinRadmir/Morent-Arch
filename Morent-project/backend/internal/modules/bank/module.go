package bank

import (
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	"morent-backend/internal/messaging"
	"morent-backend/internal/service"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *service.BankService
	Handler *handlers.BankHandler
}

func NewModule(cfg *config.Config) Outputs {
	_ = cfg
	gateway := messaging.BankNoopGateway{}
	bankService := service.NewBankService(gateway)
	handler := handlers.NewBankHandler(bankService, cfg)
	return Outputs{
		Service: bankService,
		Handler: handler,
	}
}
