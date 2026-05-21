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

	morentAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return common.WrapCORS(common.WithAuth(container.AuthService, container.Config, h))
	}

	groups := []server.RouteGroup{
		{Prefix: "/bank", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/session", Handler: morentAuth(container.BankHandler.SyncSession)},
			{Method: http.MethodPost, Path: "/auth/logout", Handler: bankAuth(container.BankHandler.Logout)},
			{Method: http.MethodGet, Path: "/profile", Handler: bankAuth(container.BankHandler.Profile)},
			{Method: http.MethodPost, Path: "/deposit", Handler: bankAuth(container.BankHandler.Deposit)},
			{Method: http.MethodPost, Path: "/transfer", Handler: bankAuth(container.BankHandler.Transfer)},
			{Method: http.MethodGet, Path: "/transactions", Handler: bankAuth(container.BankHandler.Transactions)},
		}},
		{Prefix: "/Bank", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/Session", Handler: morentAuth(container.BankHandler.SyncSession)},
			{Method: http.MethodPost, Path: "/Auth/Logout", Handler: bankAuth(container.BankHandler.Logout)},
			{Method: http.MethodGet, Path: "/Profile", Handler: bankAuth(container.BankHandler.Profile)},
			{Method: http.MethodPost, Path: "/Deposit", Handler: bankAuth(container.BankHandler.Deposit)},
			{Method: http.MethodPost, Path: "/Transfer", Handler: bankAuth(container.BankHandler.Transfer)},
			{Method: http.MethodGet, Path: "/Transactions", Handler: bankAuth(container.BankHandler.Transactions)},
		}},
	}

	server.RegisterRouteGroupsToMux(mux, groups, common.WrapCORS)
}
