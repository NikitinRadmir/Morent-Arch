package di

import (
	"context"
	"log/slog"
	"time"

	"morent-backend/internal/cache"
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	applog "morent-backend/internal/logger"
	adminapp "morent-backend/internal/modules/admin/app"
	adminmodule "morent-backend/internal/modules/admin"
	authmodule "morent-backend/internal/modules/auth"
	carsmodule "morent-backend/internal/modules/cars"
	commentsmodule "morent-backend/internal/modules/comments"
	favoritesmodule "morent-backend/internal/modules/favorites"
	rentalsmodule "morent-backend/internal/modules/rentals"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"
	"morent-backend/internal/storage"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(
		applog.New,
		provideConfig,
		config.ConnectDB,
		cache.ConnectRedis,
		provideCarCache,
		storage.NewMinioStorage,
		service.NewLogService,
		repository.NewCarRepository,
		repository.NewCommentRepository,
		repository.NewUserRepository,
		repository.NewFavoriteRepository,
		repository.NewSessionRepository,
		repository.NewRentalRepository,
		carsmodule.NewModule,
		authmodule.NewModule,
		commentsmodule.NewModule,
		favoritesmodule.NewModule,
		rentalsmodule.NewModule,
		adminmodule.NewModule,
		handlers.NewMediaHandler,
		buildContainer,
	),
	fx.Invoke(setSlogDefault),
	fx.Invoke(registerRedisHook),
)

func provideCarCache(cfg *config.Config, rdb *redis.Client) *cache.CarCache {
	return cache.NewCarCache(rdb, time.Duration(cfg.CarsCacheTTLSeconds)*time.Second)
}

func setSlogDefault(l *slog.Logger) {
	slog.SetDefault(l)
}

func registerRedisHook(lc fx.Lifecycle, rdb *redis.Client) {
	if rdb == nil {
		return
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return rdb.Close()
		},
	})
}

func provideConfig() (*config.Config, error) {
	return config.LoadFromEnv()
}

type containerParams struct {
	fx.In
	Config             *config.Config
	DB                 *gorm.DB
	CarRepository      *repository.CarRepository
	CommentRepository  *repository.CommentRepository
	FavoriteRepository *repository.FavoriteRepository
	SessionRepository  *repository.SessionRepository
	RentalRepository   *repository.RentalRepository
	UserRepository     *repository.UserRepository
	CarService         *service.CarService
	CommentService     *service.CommentService
	FavoriteService    *service.FavoriteService
	RentalService      *service.RentalService
	AuthService        *service.AuthService
	AdminService       *adminapp.Service
	Storage            *storage.MinioStorage
	LogService         *service.LogService
	AdminHandler       *handlers.AdminHandler
	MediaHandler       *handlers.MediaHandler
	CarHandler         *handlers.CarHandler
	CommentHandler     *handlers.CommentHandler
	AuthHandler        *handlers.AuthHandler
	FavoriteHandler    *handlers.FavoriteHandler
	RentalHandler      *handlers.RentalHandler
}

func buildContainer(p containerParams) *Container {
	return &Container{
		Config:             p.Config,
		DB:                 p.DB,
		CarRepository:      p.CarRepository,
		CommentRepository:  p.CommentRepository,
		FavoriteRepository: p.FavoriteRepository,
		SessionRepository:  p.SessionRepository,
		RentalRepository:   p.RentalRepository,
		UserRepository:     p.UserRepository,
		CarService:         p.CarService,
		CommentService:     p.CommentService,
		FavoriteService:    p.FavoriteService,
		RentalService:      p.RentalService,
		AuthService:        p.AuthService,
		AdminService:       p.AdminService,
		Storage:            p.Storage,
		LogService:         p.LogService,
		AdminHandler:       p.AdminHandler,
		MediaHandler:       p.MediaHandler,
		CarHandler:         p.CarHandler,
		CommentHandler:     p.CommentHandler,
		AuthHandler:        p.AuthHandler,
		FavoriteHandler:    p.FavoriteHandler,
		RentalHandler:      p.RentalHandler,
	}
}
