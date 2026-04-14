package common

import (
	"net/http"
	"strings"

	"morent-backend/internal/service"
)

func WrapCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, SOAPAction")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

func WithAdmin(auth *service.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(r.Header.Get("Authorization"))
		lower := strings.ToLower(token)
		if strings.HasPrefix(lower, "bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		user, err := auth.GetUserByToken(token)
		if err != nil || user == nil || strings.ToLower(strings.TrimSpace(user.Role)) != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
