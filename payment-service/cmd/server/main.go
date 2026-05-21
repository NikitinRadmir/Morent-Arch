package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"morent-arch/payment-service/internal/config"
	"morent-arch/payment-service/internal/handlers"
	"morent-arch/payment-service/internal/messaging"
)

func main() {
	cfg := config.Load()
	configureLogger(cfg)

	handler := handlers.NewRouter()
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var bankKafka *messaging.BankKafka
	if cfg.KafkaEnabled {
		k, err := messaging.NewBankKafka(cfg, slog.Default())
		if err != nil {
			slog.Error("kafka_init_failed", "error", err)
			os.Exit(1)
		}
		bankKafka = k
		go func() {
			if err := bankKafka.Run(ctx); err != nil {
				slog.Error("kafka_consumer_failed", "error", err)
			}
		}()
	}

	go func() {
		slog.Info("server_starting", "addr", cfg.HTTPAddr, "env", cfg.AppEnv, "kafka_enabled", cfg.KafkaEnabled)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server_shutdown_failed", "error", err)
	}
	if bankKafka != nil {
		if err := bankKafka.Close(); err != nil {
			slog.Error("kafka_close_failed", "error", err)
		}
	}
	slog.Info("server_stopped")
}

func configureLogger(cfg config.Config) {
	level := new(slog.LevelVar)
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "warn":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		level.Set(slog.LevelInfo)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
}
