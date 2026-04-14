package auth

import (
	"morent-backend/internal/handlers"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *service.AuthService
	Handler *handlers.AuthHandler
}

func NewModule(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	logService *service.LogService,
) Outputs {
	authService := service.NewAuthService(userRepo, sessionRepo)
	authHandler := handlers.NewAuthHandler(authService, logService)

	return Outputs{
		Service: authService,
		Handler: authHandler,
	}
}
