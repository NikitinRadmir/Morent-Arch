package admin

import (
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	"morent-backend/internal/messaging"
	adminapp "morent-backend/internal/modules/admin/app"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"
	"morent-backend/internal/storage"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *adminapp.Service
	Handler *handlers.AdminHandler
}

func NewModule(
	userRepo *repository.UserRepository,
	rentalRepo *repository.RentalRepository,
	favoriteRepo *repository.FavoriteRepository,
	commentRepo *repository.CommentRepository,
	logService *service.LogService,
	carService *service.CarService,
	cfg *config.Config,
	minioStorage *storage.MinioStorage,
	events messaging.UserEventPublisher,
) Outputs {
	adminService := adminapp.NewService(
		userRepo, rentalRepo, favoriteRepo, commentRepo, logService,
		events, cfg.MorentCompanyName, nil,
	)
	adminHandler := handlers.NewAdminHandler(adminService, logService, carService, cfg, minioStorage)

	return Outputs{
		Service: adminService,
		Handler: adminHandler,
	}
}
