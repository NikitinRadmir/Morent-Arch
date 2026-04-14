package rentals

import (
	"morent-backend/internal/handlers"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *service.RentalService
	Handler *handlers.RentalHandler
}

func NewModule(
	authService *service.AuthService,
	logService *service.LogService,
	rentalRepo *repository.RentalRepository,
	carRepo *repository.CarRepository,
) Outputs {
	rentalService := service.NewRentalService(rentalRepo, carRepo)
	rentalHandler := handlers.NewRentalHandler(authService, rentalService, logService)

	return Outputs{
		Service: rentalService,
		Handler: rentalHandler,
	}
}
