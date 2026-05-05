package grpctransport

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"morent-backend/internal/di"
	grpcapi "morent-backend/internal/grpc"
)

func RegisterLifecycle(lc fx.Lifecycle, container *di.Container) {
	var grpcSrv *grpcapi.Server
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var grpcStartErr error
			grpcSrv, grpcStartErr = grpcapi.Start(container.Config, container.RentalService)
			if grpcStartErr != nil {
				return grpcStartErr
			}
			grpcPort := container.Config.GRPCPort
			if grpcPort == "" {
				grpcPort = "50051"
			}
			log.Printf("gRPC сервер запущен на порту :%s", grpcPort)
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
