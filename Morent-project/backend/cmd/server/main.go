package main

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"morent-backend/internal/di"
	gqlschema "morent-backend/internal/graphql"
	"morent-backend/internal/migrations"
	grpctransport "morent-backend/internal/modules/transport/grpc"
	httptransport "morent-backend/internal/modules/transport/http"
	"morent-backend/internal/seed"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	app := fx.New(
		di.Module,
		fx.Provide(gqlschema.NewSchema),
		fx.Invoke(registerDatabaseLifecycle),
		fx.Invoke(httptransport.RegisterLifecycle),
		fx.Invoke(grpctransport.RegisterLifecycle),
	)
	app.Run()
}

func registerDatabaseLifecycle(lc fx.Lifecycle, container *di.Container) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sqlDB, err := container.DB.DB()
			if err != nil {
				return err
			}
			if err := sqlDB.PingContext(ctx); err != nil {
				return err
			}
			log.Println("Подключение к БД установлено")

			if err := migrations.RunMigrations(container.DB); err != nil {
				return err
			}

			if len(os.Args) > 1 && os.Args[1] == "seed" {
				if err := seed.RunSeed(container.DB); err != nil {
					return err
				}
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			sqlDB, err := container.DB.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
}
