package auth

import (
	"log/slog"

	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	"morent-backend/internal/messaging"
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
	events messaging.UserEventPublisher,
	cfg *config.Config,
	log *slog.Logger,
) Outputs {
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		events,
		cfg.MorentCompanyName,
		log,
	)
	authHandler := handlers.NewAuthHandler(authService, logService)

	return Outputs{
		Service: authService,
		Handler: authHandler,
	}
}
