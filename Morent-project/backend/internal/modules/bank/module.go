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

type ModuleParams struct {
	fx.In
	Cfg     *config.Config
	Gateway messaging.BankGateway
	Auth    *service.AuthService
}

func NewModule(p ModuleParams) Outputs {
	bankService := service.NewBankService(p.Gateway, p.Cfg.BankLinkSecret)
	handler := handlers.NewBankHandler(bankService, p.Auth, p.Cfg)
	return Outputs{
		Service: bankService,
		Handler: handler,
	}
}
