package httptransport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"morent-backend/internal/di"
	"morent-backend/internal/health"
	admintransport "morent-backend/internal/modules/transport/http/admin"
	banktransport "morent-backend/internal/modules/transport/http/bank"
	"morent-backend/internal/modules/transport/http/common"
	graphqltransport "morent-backend/internal/modules/transport/http/graphql"
	"morent-backend/internal/modules/transport/http/public"
	"morent-backend/internal/server"

	"obslog"
)

func RegisterLifecycle(lc fx.Lifecycle, container *di.Container, schema graphql.Schema, log *slog.Logger, rdb *redis.Client) {
	if log == nil {
		log = slog.Default()
	}

	mux := http.NewServeMux()
	readiness := &health.Checker{DB: container.DB, Redis: rdb}
	mux.HandleFunc("/health", common.WrapCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodOptions {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"morent-backend"}`))
	}))
	mux.HandleFunc("/ready", common.WrapCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodOptions {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), health.DefaultCheckTimeout)
		defer cancel()
		report := readiness.Check(ctx)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(report.HTTPStatus())
		_ = json.NewEncoder(w).Encode(report)
	}))
	mux.HandleFunc("/swagger.json", common.WrapCORS(server.SwaggerJSON))
	mux.HandleFunc("/swagger", common.WrapCORS(server.SwaggerUI))
	mux.HandleFunc("/swagger/", common.WrapCORS(server.SwaggerUI))
	graphqltransport.Register(mux, schema)
	public.Register(mux, container)
	banktransport.Register(mux, container)
	admintransport.Register(mux, container)
	mux.HandleFunc("/", common.WrapCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not Found","message":"The requested resource was not found"}`))
	}))

	handler := common.WithRecover(log, common.WithAccessLog(log, mux))

	httpSrv := &http.Server{
		Addr:              ":1488",
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	heartbeatCtx, heartbeatCancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				ticker := time.NewTicker(5 * time.Minute)
				defer ticker.Stop()
				for {
					select {
					case <-heartbeatCtx.Done():
						return
					case <-ticker.C:
						obslog.Heartbeat(log, "process", "alive")
						if errLog := container.Storage.AppendDailyLog(context.Background(), "minio heartbeat (5m)"); errLog != nil {
							log.Warn("minio log write error", "log_type", obslog.LogTypeApp, "error", errLog)
						}
					}
				}
			}()

			go func() {
				log.Info("http server starting", "addr", httpSrv.Addr)
				if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("http server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			heartbeatCancel()
			return httpSrv.Shutdown(ctx)
		},
	})
}
