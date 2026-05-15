package grpctransport

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"morent-backend/internal/di"
	grpcapi "morent-backend/internal/grpc"
)

func RegisterLifecycle(lc fx.Lifecycle, container *di.Container, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	var grpcSrv *grpcapi.Server
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			grpcSrv, err = grpcapi.Start(container.Config, container.RentalService, log)
			if err != nil {
				return err
			}
			grpcPort := container.Config.GRPCPort
			if grpcPort == "" {
				grpcPort = "50051"
			}
			log.Info("grpc server started", "port", grpcPort)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if grpcSrv != nil {
				grpcSrv.Stop()
			}
			return nil
		},
	})
}
