package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.FromEnv()

	zerolog.TimeFieldFormat = time.RFC3339Nano
	if cfg.Env == "dev" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	srv := httpserver.NewServer(cfg)

	httpSrv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: srv.Router(),
	}

	// запуск HTTP-сервера
	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("http: starting server")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("http: server crashed")
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info().Msg("http: shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("http: shutdown error")
	}
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("app: shutdown error")
	}

	log.Info().Msg("bye")
}
