package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"morent-backend/internal/models"
	favoritesdto "morent-backend/internal/modules/favorites/httpdto"
	"morent-backend/internal/service"
)

type FavoriteHandler struct {
	authService     *service.AuthService
	favoriteService *service.FavoriteService
	logService      *service.LogService
}

func NewFavoriteHandler(authService *service.AuthService, favoriteService *service.FavoriteService, logService *service.LogService) *FavoriteHandler {
	return &FavoriteHandler{
		authService:     authService,
		favoriteService: favoriteService,
		logService:      logService,
	}
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	user, authErr := h.authenticate(r)
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	cars, errList := h.favoriteService.List(user.ID)
	if errList != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:    time.Now(),
			Type:    service.LogFavorite,
			Action:  "list_favorites",
			UserID:  user.ID,
			Result:  "error",
			Message: errList.Error(),
		})
		http.Error(w, "Failed to load favorites: "+errList.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogFavorite,
		Action: "list_favorites",
		UserID: user.ID,
		Result: "success",
	})
	json.NewEncoder(w).Encode(cars)
}

func (h *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	user, authErr := h.authenticate(r)
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	var payload favoritesdto.AddFavoriteRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&payload); errDecode != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if payload.CarID == 0 {
		http.Error(w, "carId is required", http.StatusBadRequest)
		return
	}

	if errAdd := h.favoriteService.Add(user.ID, payload.CarID); errAdd != nil {
		status := http.StatusInternalServerError
		if errAdd == service.ErrCarNotFound {
			status = http.StatusNotFound
		}
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogFavorite,
			Action:   "add_favorite",
			UserID:   user.ID,
			ObjectID: payload.CarID,
			Result:   "error",
			Message:  errAdd.Error(),
		})
		http.Error(w, errAdd.Error(), status)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogFavorite,
		Action:   "add_favorite",
		UserID:   user.ID,
		ObjectID: payload.CarID,
		Result:   "success",
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "added"})
}

func (h *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {
	user, authErr := h.authenticate(r)
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusUnauthorized)
		return
	}
	ctx := r.Context()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[len(parts)-1]
	carID64, errParse := strconv.ParseUint(idStr, 10, 64)
	if errParse != nil {
		http.Error(w, "invalid car id", http.StatusBadRequest)
		return
	}

	if errRemove := h.favoriteService.Remove(user.ID, uint(carID64)); errRemove != nil {
		_ = h.logService.LogEvent(ctx, service.LogEvent{
			Time:     time.Now(),
			Type:     service.LogFavorite,
			Action:   "remove_favorite",
			UserID:   user.ID,
			ObjectID: carID64,
			Result:   "error",
			Message:  errRemove.Error(),
		})
		http.Error(w, "failed to remove favorite: "+errRemove.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogFavorite,
		Action:   "remove_favorite",
		UserID:   user.ID,
		ObjectID: carID64,
		Result:   "success",
	})
	w.WriteHeader(http.StatusNoContent)
}

func (h *FavoriteHandler) authenticate(r *http.Request) (*models.User, error) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		cookieName := strings.TrimSpace(os.Getenv("SESSION_COOKIE_NAME"))
		if cookieName == "" {
			cookieName = "morent_session"
		}
		cookie, cookieErr := r.Cookie(cookieName)
		if cookieErr == nil && cookie != nil {
			token = strings.TrimSpace(cookie.Value)
		}
	}
	if token == "" {
		return nil, service.ErrInvalidToken
	}
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		token = strings.TrimSpace(token[7:])
	}
	user, getUserByTokenErr := h.authService.GetUserByToken(token)
	if getUserByTokenErr != nil {
		return nil, getUserByTokenErr
	}
	return user, nil
}
