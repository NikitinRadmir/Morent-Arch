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

	"morent-arch/payment-service/internal/app"
	"morent-arch/payment-service/internal/config"
	"morent-arch/payment-service/internal/handlers"
	"morent-arch/payment-service/internal/messaging"

	"obslog"
)

func main() {
	cfg := config.Load()
	configureLogger(cfg)

	deps, err := app.Bootstrap(cfg)
	if err != nil {
		slog.Error("bootstrap_failed", "error", err)
		os.Exit(1)
	}

	handler := handlers.NewRouter(deps.Payment, deps.DB)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var bankKafka *messaging.BankKafka
	if cfg.KafkaEnabled {
		k, err := messaging.NewBankKafka(cfg, deps.Processor, slog.Default())
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
	logger := obslog.New("payment-service", obslog.LogTypeApp)
	slog.SetDefault(logger)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			obslog.Heartbeat(logger, "process", "alive")
		}
	}()
}
