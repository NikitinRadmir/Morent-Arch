package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"morent-backend/internal/config"
	"morent-backend/internal/models"
	favoritesdto "morent-backend/internal/modules/favorites/httpdto"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/service"
)

type FavoriteHandler struct {
	authService     *service.AuthService
	favoriteService *service.FavoriteService
	logService      *service.LogService
	cfg             *config.Config
}

func NewFavoriteHandler(authService *service.AuthService, favoriteService *service.FavoriteService, logService *service.LogService, cfg *config.Config) *FavoriteHandler {
	return &FavoriteHandler{
		authService:     authService,
		favoriteService: favoriteService,
		logService:      logService,
		cfg:             cfg,
	}
}

func (h *FavoriteHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.authenticate(r)
	if err != nil {
		common.WriteUnauthorized(w, "Войдите в аккаунт")
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
		common.WriteInternalErrorJSON(w, "Не удалось загрузить избранное")
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
	user, err := h.authenticate(r)
	if err != nil {
		common.WriteUnauthorized(w, "Войдите в аккаунт")
		return
	}
	ctx := r.Context()

	var payload favoritesdto.AddFavoriteRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&payload); errDecode != nil {
		common.WriteBadRequest(w, "Некорректный формат запроса", "invalid_body")
		return
	}
	if payload.CarID == 0 {
		common.WriteBadRequest(w, "Выберите автомобиль", "car_required")
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
		if status == http.StatusNotFound {
			common.WriteNotFound(w, "Автомобиль не найден")
			return
		}
		common.WriteInternalErrorJSON(w, "Не удалось добавить автомобиль в избранное")
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
	user, err := h.authenticate(r)
	if err != nil {
		common.WriteUnauthorized(w, "Войдите в аккаунт")
		return
	}
	ctx := r.Context()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		common.WriteBadRequest(w, "Некорректный путь запроса", "invalid_path")
		return
	}
	idStr := parts[len(parts)-1]
	carID64, errParse := strconv.ParseUint(idStr, 10, 64)
	if errParse != nil {
		common.WriteBadRequest(w, "Некорректный идентификатор автомобиля", "invalid_car_id")
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
		common.WriteInternalErrorJSON(w, "Не удалось убрать автомобиль из избранного")
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
	return common.Authenticate(h.authService, h.cfg, r)
}
