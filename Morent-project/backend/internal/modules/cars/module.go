package cars

import (
	"morent-backend/internal/handlers"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *service.CarService
	Handler *handlers.CarHandler
}

func NewModule(carRepo *repository.CarRepository, logService *service.LogService) Outputs {
	carService := service.NewCarService(carRepo)
	carHandler := handlers.NewCarHandler(carService, logService)

	return Outputs{
		Service: carService,
		Handler: carHandler,
	}
}
