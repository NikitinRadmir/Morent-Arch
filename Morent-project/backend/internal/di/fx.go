package di

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"morent-backend/internal/cache"
	"morent-backend/internal/config"
	"morent-backend/internal/handlers"
	"morent-backend/internal/messaging"
	kafkamsg "morent-backend/internal/messaging/kafka"
	applog "morent-backend/internal/logger"
	adminapp "morent-backend/internal/modules/admin/app"
	adminmodule "morent-backend/internal/modules/admin"
	authmodule "morent-backend/internal/modules/auth"
	bankmodule "morent-backend/internal/modules/bank"
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
		bankmodule.NewModule,
		commentsmodule.NewModule,
		favoritesmodule.NewModule,
		rentalsmodule.NewModule,
		adminmodule.NewModule,
		handlers.NewMediaHandler,
		provideGeneratorClient,
		provideUserEventPublisher,
		provideEmailEventPublisher,
		provideEmailNotifier,
		provideBankGateway,
		handlers.NewPasswordHandler,
		buildContainer,
	),
	fx.Invoke(setSlogDefault),
	fx.Invoke(registerRedisHook),
	fx.Invoke(registerKafkaHook),
	fx.Invoke(registerBankKafkaHook),
	fx.Invoke(registerRentalReminderHook),
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

func provideGeneratorClient(cfg *config.Config) *service.GeneratorClient {
	return service.NewGeneratorClient(cfg.GeneratorBaseURL)
}

func provideUserEventPublisher(cfg *config.Config, log *slog.Logger) messaging.UserEventPublisher {
	if !cfg.KafkaEnabled {
		return messaging.NoopPublisher{}
	}
	pub, err := kafkamsg.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopicUsers, cfg.MorentCompanyName, log)
	if err != nil {
		log.Warn("kafka publisher disabled", "error", err)
		return messaging.NoopPublisher{}
	}
	return pub
}

func provideEmailNotifier(
	cfg *config.Config,
	users *repository.UserRepository,
	emails messaging.EmailEventPublisher,
	log *slog.Logger,
) *service.EmailNotifier {
	return service.NewEmailNotifier(users, emails, cfg.FrontendOrigin, log)
}

func provideEmailEventPublisher(cfg *config.Config, log *slog.Logger) messaging.EmailEventPublisher {
	if !cfg.KafkaEnabled {
		return messaging.EmailNoopPublisher{}
	}
	pub, err := kafkamsg.NewEmailPublisher(cfg.KafkaBrokers, cfg.KafkaTopicEmails, log)
	if err != nil {
		log.Warn("kafka email publisher disabled", "error", err)
		return messaging.EmailNoopPublisher{}
	}
	log.Info("kafka email publisher enabled", "topic", cfg.KafkaTopicEmails, "brokers", cfg.KafkaBrokers)
	return pub
}

func provideBankGateway(cfg *config.Config, log *slog.Logger) messaging.BankGateway {
	if !cfg.KafkaEnabled || strings.TrimSpace(cfg.KafkaBrokers) == "" {
		return messaging.BankNoopGateway{}
	}
	gw, err := kafkamsg.NewBankGateway(
		cfg.KafkaBrokers,
		cfg.KafkaTopicBankCommands,
		cfg.KafkaTopicBankResponses,
		cfg.KafkaGroupMorentBank,
		log,
	)
	if err != nil {
		log.Warn("bank kafka gateway disabled", "error", err)
		return messaging.BankNoopGateway{}
	}
	return gw
}

func registerRentalReminderHook(
	lc fx.Lifecycle,
	rentals *repository.RentalRepository,
	users *repository.UserRepository,
	emails *service.EmailNotifier,
	log *slog.Logger,
) {
	scheduler := service.NewRentalReminderScheduler(rentals, users, emails, log)
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go scheduler.Run(ctx)
			log.Info("rental day reminder scheduler started")
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}

func registerBankKafkaHook(lc fx.Lifecycle, cfg *config.Config, gw messaging.BankGateway, log *slog.Logger) {
	if !cfg.KafkaEnabled {
		return
	}
	runner, ok := gw.(interface {
		Run(context.Context) error
		Close() error
	})
	if !ok {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := runner.Run(ctx); err != nil && ctx.Err() == nil {
					log.Warn("bank kafka consumer stopped", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return runner.Close()
		},
	})
}

func registerKafkaHook(lc fx.Lifecycle, cfg *config.Config, pub messaging.UserEventPublisher, emailPub messaging.EmailEventPublisher) {
	if !cfg.KafkaEnabled {
		return
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			var err error
			if closer, ok := pub.(interface{ Close() error }); ok {
				if e := closer.Close(); e != nil {
					err = e
				}
			}
			if closer, ok := emailPub.(interface{ Close() error }); ok {
				if e := closer.Close(); e != nil && err == nil {
					err = e
				}
			}
			return err
		},
	})
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
	PasswordHandler    *handlers.PasswordHandler
	BankService        *service.BankService
	BankHandler        *handlers.BankHandler
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
		PasswordHandler:    p.PasswordHandler,
		BankService:        p.BankService,
		BankHandler:        p.BankHandler,
	}
}
