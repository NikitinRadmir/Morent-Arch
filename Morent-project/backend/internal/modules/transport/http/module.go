package httptransport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"morent-backend/internal/di"
	admintransport "morent-backend/internal/modules/transport/http/admin"
	"morent-backend/internal/modules/transport/http/common"
	graphqltransport "morent-backend/internal/modules/transport/http/graphql"
	"morent-backend/internal/modules/transport/http/public"
	"morent-backend/internal/server"
)

func RegisterLifecycle(lc fx.Lifecycle, container *di.Container, schema graphql.Schema) {
	mux := http.NewServeMux()
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
	mux.HandleFunc("/swagger.json", common.WrapCORS(server.SwaggerJSON))
	mux.HandleFunc("/swagger", common.WrapCORS(server.SwaggerUI))
	mux.HandleFunc("/swagger/", common.WrapCORS(server.SwaggerUI))
	graphqltransport.Register(mux, schema)
	public.Register(mux, container)
	admintransport.Register(mux, container)
	mux.HandleFunc("/", common.WrapCORS(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Not Found","message":"The requested resource was not found"}`))
	}))

	httpPort := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if httpPort == "" {
		httpPort = strings.TrimSpace(os.Getenv("BACKEND_PORT"))
	}
	if httpPort == "" {
		httpPort = "1488"
	}

	httpSrv := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           mux,
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
						if errLog := container.Storage.AppendDailyLog(context.Background(), "minio heartbeat (5m)"); errLog != nil {
							log.Println("minio log write error:", errLog)
						}
					}
				}
			}()

			go func() {
				fmt.Printf("HTTP сервер запущен на порту %s\n", httpSrv.Addr)
				fmt.Print("http://localhost:" + httpPort + "/")
				if listenErr := httpSrv.ListenAndServe(); listenErr != nil && listenErr != http.ErrServerClosed {
					log.Error("Ошибка запуска HTTP сервера:", listenErr)
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
