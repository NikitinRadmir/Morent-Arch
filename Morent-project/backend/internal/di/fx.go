package di

import (
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
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

	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(
		provideConfig,
		config.ConnectDB,
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
)

func provideConfig() (*config.Config, error) {
	return config.LoadConfig("config.json")
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
