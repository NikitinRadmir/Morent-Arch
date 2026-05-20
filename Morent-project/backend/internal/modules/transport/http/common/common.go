package common

import (
	"net/http"
	"os"
	"strings"

	"morent-backend/internal/config"
	"morent-backend/internal/service"
)

func WrapCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		allowedOrigin := strings.TrimSpace(os.Getenv("FRONTEND_ORIGIN"))
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:5173"
		}
		if origin != "" && (origin == allowedOrigin || origin == "http://localhost:1488") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, SOAPAction")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

func WithAdmin(auth *service.AuthService, cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := Authenticate(auth, cfg, r)
		if err != nil || user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if strings.ToLower(strings.TrimSpace(user.Role)) != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
