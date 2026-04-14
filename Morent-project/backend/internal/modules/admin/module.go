package admin

import (
	"morent-backend/internal/handlers"
	adminapp "morent-backend/internal/modules/admin/app"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"

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
) Outputs {
	adminService := adminapp.NewService(userRepo, rentalRepo, favoriteRepo, commentRepo, logService)
	adminHandler := handlers.NewAdminHandler(adminService, logService, carService)

	return Outputs{
		Service: adminService,
		Handler: adminHandler,
	}
}
