package di

import (
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	adminapp "morent-backend/internal/modules/admin/app"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"
	"morent-backend/internal/storage"

	"gorm.io/gorm"
)

type Container struct {
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
	PasswordHandler    *handlers.PasswordHandler
}
