package grpcapi

import (
	"fmt"
	"log/slog"
	"net"

	"morent-backend/internal/config"
	"morent-backend/internal/grpc/bookingpb/booking/v1"
	"morent-backend/internal/service"

	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func Start(cfg *config.Config, rentalService *service.RentalService, log *slog.Logger) (*Server, error) {
	port := cfg.GRPCPort
	if port == "" {
		port = "50051"
	}
	addr := fmt.Sprintf(":%s", port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	if log == nil {
		log = slog.Default()
	}
	s := grpc.NewServer(grpc.ChainUnaryInterceptor(recoveryUnaryInterceptor(log)))
	bookingpb.RegisterBookingServiceServer(s, NewBookingServiceServer(rentalService))

	go func() {
		_ = s.Serve(lis)
	}()

	return &Server{
		grpcServer: s,
		listener:   lis,
	}, nil
}

func (s *Server) Stop() {
	if s == nil {
		return
	}
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
}

