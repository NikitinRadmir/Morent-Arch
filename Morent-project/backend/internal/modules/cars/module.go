package cars

import (
	"log/slog"

	"morent-backend/internal/cache"
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

func NewModule(carRepo *repository.CarRepository, logService *service.LogService, carCache *cache.CarCache, log *slog.Logger) Outputs {
	carService := service.NewCarService(carRepo, carCache, log)
	carHandler := handlers.NewCarHandler(carService, logService)

	return Outputs{
		Service: carService,
		Handler: carHandler,
	}
}
