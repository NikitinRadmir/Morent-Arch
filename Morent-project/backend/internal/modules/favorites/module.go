package favorites

import (
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *service.FavoriteService
	Handler *handlers.FavoriteHandler
}

func NewModule(
	favoriteRepo *repository.FavoriteRepository,
	carRepo *repository.CarRepository,
	authService *service.AuthService,
	logService *service.LogService,
	cfg *config.Config,
) Outputs {
	favoriteService := service.NewFavoriteService(favoriteRepo, carRepo)
	favoriteHandler := handlers.NewFavoriteHandler(authService, favoriteService, logService, cfg)

	return Outputs{
		Service: favoriteService,
		Handler: favoriteHandler,
	}
}
