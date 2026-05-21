package bank

import (
	"net/http"

	"morent-backend/internal/di"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/server"
)

func Register(mux *http.ServeMux, container *di.Container) {
	bankAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return common.WrapCORS(common.WithBankAuth(container.BankService, container.Config, h))
	}

	groups := []server.RouteGroup{
		{Prefix: "/bank", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/auth/register", Handler: container.BankHandler.Register},
			{Method: http.MethodPost, Path: "/auth/login", Handler: container.BankHandler.Login},
			{Method: http.MethodPost, Path: "/auth/logout", Handler: bankAuth(container.BankHandler.Logout)},
			{Method: http.MethodGet, Path: "/profile", Handler: bankAuth(container.BankHandler.Profile)},
			{Method: http.MethodPost, Path: "/deposit", Handler: bankAuth(container.BankHandler.Deposit)},
			{Method: http.MethodPost, Path: "/transfer", Handler: bankAuth(container.BankHandler.Transfer)},
			{Method: http.MethodGet, Path: "/transactions", Handler: bankAuth(container.BankHandler.Transactions)},
		}},
		{Prefix: "/Bank", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/Auth/Register", Handler: container.BankHandler.Register},
			{Method: http.MethodPost, Path: "/Auth/Login", Handler: container.BankHandler.Login},
			{Method: http.MethodPost, Path: "/Auth/Logout", Handler: bankAuth(container.BankHandler.Logout)},
			{Method: http.MethodGet, Path: "/Profile", Handler: bankAuth(container.BankHandler.Profile)},
			{Method: http.MethodPost, Path: "/Deposit", Handler: bankAuth(container.BankHandler.Deposit)},
			{Method: http.MethodPost, Path: "/Transfer", Handler: bankAuth(container.BankHandler.Transfer)},
			{Method: http.MethodGet, Path: "/Transactions", Handler: bankAuth(container.BankHandler.Transactions)},
		}},
	}

	server.RegisterRouteGroupsToMux(mux, groups, common.WrapCORS)
}
