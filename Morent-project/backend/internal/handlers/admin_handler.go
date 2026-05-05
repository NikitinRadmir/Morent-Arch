package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"morent-backend/internal/config"
	"morent-backend/internal/models"
	adminapp "morent-backend/internal/modules/admin/app"
	admindto "morent-backend/internal/modules/admin/httpdto"
	"morent-backend/internal/service"
	"morent-backend/internal/storage"
)

type AdminHandler struct {
	adminService  *adminapp.Service
	logService    *service.LogService
	carService    *service.CarService
	cfg           *config.Config
	storage       *storage.MinioStorage
	httpClient    *http.Client
	aggregatorURL string
	retryCount    int
}

func NewAdminHandler(
	adminService *adminapp.Service,
	logService *service.LogService,
	carService *service.CarService,
	cfg *config.Config,
	storage *storage.MinioStorage,
) *AdminHandler {
	baseURL := strings.TrimRight(os.Getenv("AGGREGATOR_BASE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	timeoutSec := cfg.AggregatorTimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 35
	}
	retryCount := cfg.AggregatorRetryCount
	if retryCount <= 0 {
		retryCount = 3
	}

	return &AdminHandler{
		adminService: adminService,
		logService:   logService,
		carService:   carService,
		cfg:          cfg,
		storage:      storage,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
		aggregatorURL: baseURL,
		retryCount:    retryCount,
	}
}

// ---- Users ----

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, listUsersErr := h.adminService.ListUsers()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_users",
		Result: "success",
	}
	if listUsersErr != nil {
		logEvent.Result = "error"
		logEvent.Message = listUsersErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		http.Error(w, listUsersErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req admindto.UpdateUserRequest
	if decodeUserReqErr := json.NewDecoder(r.Body).Decode(&req); decodeUserReqErr != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	u := req.User
	if u.ID == 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	updateUserErr := h.adminService.UpdateUser(&u)
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		UserID: u.ID,
		Action: "update_user",
		Data:   u,
		Result: "success",
	}
	if updateUserErr != nil {
		logEvent.Result = "error"
		logEvent.Message = updateUserErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(updateUserErr, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, updateUserErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(u)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	deleteUserErr := h.adminService.DeleteUser(id)
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		UserID: id,
		Action: "delete_user",
		Result: "success",
	}
	if deleteUserErr != nil {
		logEvent.Result = "error"
		logEvent.Message = deleteUserErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(deleteUserErr, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, deleteUserErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Rentals ----

func (h *AdminHandler) ListRentals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rents, listRentalsErr := h.adminService.ListRentals()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_rentals",
		Result: "success",
	}
	if listRentalsErr != nil {
		logEvent.Result = "error"
		logEvent.Message = listRentalsErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		http.Error(w, listRentalsErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(rents)
}

func (h *AdminHandler) DeleteRental(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	deleteRentalErr := h.adminService.DeleteRental(uint(id))
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		UserID:   nil,
		ObjectID: id,
		Action:   "delete_rental",
		Result:   "success",
	}
	if deleteRentalErr != nil {
		logEvent.Result = "error"
		logEvent.Message = deleteRentalErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(deleteRentalErr, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, deleteRentalErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Favorites ----

func (h *AdminHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	favs, listFavoritesErr := h.adminService.ListFavorites()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_favorites",
		Result: "success",
	}
	if listFavoritesErr != nil {
		logEvent.Result = "error"
		logEvent.Message = listFavoritesErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		http.Error(w, listFavoritesErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(favs)
}

func (h *AdminHandler) DeleteFavorite(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		http.Error(w, "invalid path, expected /Admin/Favorites/{userId}/{carId}", http.StatusBadRequest)
		return
	}
	userID, err1 := strconv.ParseUint(parts[len(parts)-2], 10, 64)
	carID, err2 := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "invalid ids", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	deleteFavoriteErr := h.adminService.DeleteFavorite(uint(userID), uint(carID))
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		UserID:   userID,
		ObjectID: carID,
		Action:   "delete_favorite",
		Result:   "success",
	}
	if deleteFavoriteErr != nil {
		logEvent.Result = "error"
		logEvent.Message = deleteFavoriteErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		http.Error(w, deleteFavoriteErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Comments ----

func (h *AdminHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	comments, listCommentsErr := h.adminService.ListComments()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_comments",
		Result: "success",
	}
	if listCommentsErr != nil {
		logEvent.Result = "error"
		logEvent.Message = listCommentsErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		http.Error(w, listCommentsErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(comments)
}

func (h *AdminHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	var req admindto.UpdateCommentRequest
	if decodeCommentReqErr := json.NewDecoder(r.Body).Decode(&req); decodeCommentReqErr != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	c := req.Comment
	if c.ID == 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	updateCommentErr := h.adminService.UpdateComment(&c)
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		ObjectID: c.ID,
		Action:   "update_comment",
		Data:     c,
		Result:   "success",
	}
	if updateCommentErr != nil {
		logEvent.Result = "error"
		logEvent.Message = updateCommentErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(updateCommentErr, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, updateCommentErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(c)
}

func (h *AdminHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	deleteCommentErr := h.adminService.DeleteComment(uint(id))
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		ObjectID: id,
		Action:   "delete_comment",
		Result:   "success",
	}
	if deleteCommentErr != nil {
		logEvent.Result = "error"
		logEvent.Message = deleteCommentErr.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(deleteCommentErr, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, deleteCommentErr.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ListLogs возвращает список событий аудита за сегодняшний день.
// Используется в админке для просмотра истории изменений.
func (h *AdminHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	events, listLogsErr := h.adminService.ListLogs(ctx, time.Now())
	if listLogsErr != nil {
		http.Error(w, "failed to load logs: "+listLogsErr.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(events)
}

type AggregatorTrim struct {
	ID           int     `json:"id"`
	Year         int     `json:"year"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Trim         string  `json:"trim"`
	Description  string  `json:"description"`
	MSRP         int     `json:"msrp"`
	Transmission string  `json:"transmission"`
	Seats        int     `json:"seats"`
	Fuel         float64 `json:"fuel"`
	ImageURL     string  `json:"imageUrl"`
}

type AggregatorSearchResponse struct {
	Query string           `json:"query"`
	Count int              `json:"count"`
	Cars  []AggregatorTrim `json:"cars"`
}

type ImportAggregatorCarRequest struct {
	Trim AggregatorTrim `json:"trim"`
}

func (h *AdminHandler) ListAggregatorCars(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		http.Error(w, "q is required", http.StatusBadRequest)
		return
	}

	body, marshalAggregatorSearchReqErr := json.Marshal(map[string]string{"q": query})
	if marshalAggregatorSearchReqErr != nil {
		http.Error(w, "failed to prepare request", http.StatusInternalServerError)
		return
	}

	aggregatorReq, newAggregatorReqErr := http.NewRequestWithContext(r.Context(), http.MethodPost, h.aggregatorURL+"/search/trims", bytes.NewReader(body))
	if newAggregatorReqErr != nil {
		http.Error(w, "failed to prepare aggregator request", http.StatusInternalServerError)
		return
	}
	aggregatorReq.Header.Set("Content-Type", "application/json")
	aggregatorReq.Header.Set("Accept", "application/json")

	resp, aggregatorResponseErr := h.doAggregatorRequest(aggregatorReq)
	if aggregatorResponseErr != nil {
		http.Error(w, "Сервис агрегатора недоступен", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		http.Error(w, "failed to read aggregator response", http.StatusBadGateway)
		return
	}

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusBadGateway)
		if len(respBody) == 0 {
			w.Write([]byte("aggregator request failed"))
			return
		}
		w.Write(respBody)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (h *AdminHandler) ImportAggregatorCar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ImportAggregatorCarRequest
	if decodeImportErr := json.NewDecoder(r.Body).Decode(&req); decodeImportErr != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Trim.Model) == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	nameParts := []string{}
	if req.Trim.Year > 0 {
		nameParts = append(nameParts, strconv.Itoa(req.Trim.Year))
	}
	if req.Trim.Make != "" {
		nameParts = append(nameParts, req.Trim.Make)
	}
	if req.Trim.Model != "" {
		nameParts = append(nameParts, req.Trim.Model)
	}
	if req.Trim.Trim != "" {
		nameParts = append(nameParts, req.Trim.Trim)
	}
	carName := strings.TrimSpace(strings.Join(nameParts, " "))
	if carName == "" {
		carName = fmt.Sprintf("%s %s", req.Trim.Make, req.Trim.Model)
	}

	transmission := strings.TrimSpace(req.Trim.Transmission)
	if transmission == "" {
		transmission = "Manual"
		descLower := strings.ToLower(req.Trim.Description)
		if strings.Contains(descLower, " auto") || strings.Contains(descLower, "cvt") || strings.Contains(req.Trim.Description, "A)") {
			transmission = "Automatic"
		}
	}

	price := 60.0
	if req.Trim.MSRP > 0 {
		derived := (float64(req.Trim.MSRP) * 0.01) / 2.0
		if derived < 30 {
			derived = 30
		}
		if derived > 500 {
			derived = 500
		}
		price = derived
	}

	car := models.Car{
		Name:         carName,
		Type:         "Sedan",
		Capacity:     4,
		Price:        price,
		Fuel:         req.Trim.Fuel,
		Transmission: transmission,
		ImgSrc:       req.Trim.ImageURL,
		Description:  req.Trim.Description,
	}
	if req.Trim.Seats > 0 {
		car.Capacity = req.Trim.Seats
	}
	if req.Trim.ImageURL != "" {
		if uploadedURL, uploadImageErr := h.uploadAggregatorImage(r.Context(), req.Trim.ImageURL, req.Trim.Make, req.Trim.Model); uploadImageErr == nil && uploadedURL != "" {
			car.ImgSrc = uploadedURL
		}
	}

	if createCarErr := h.carService.Create(&car); createCarErr != nil {
		http.Error(w, "failed to import car: "+createCarErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(car)
}

// ---- Helpers ----

func parseIDFromPath(path string) (uint64, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, false
	}
	id, parseIDErr := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	if parseIDErr != nil {
		return 0, false
	}
	return id, true
}

func (h *AdminHandler) doAggregatorRequest(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 1; attempt <= h.retryCount; attempt++ {
		cloned := req.Clone(req.Context())
		resp, httpDoErr := h.httpClient.Do(cloned)
		if httpDoErr == nil {
			return resp, nil
		}
		lastErr = httpDoErr
		sleepMs := int(math.Min(float64(200*attempt), 1200))
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
	return nil, lastErr
}

func (h *AdminHandler) uploadAggregatorImage(ctx context.Context, imageURL string, make string, model string) (string, error) {
	httpReq, httpReqErr := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if httpReqErr != nil {
		return "", httpReqErr
	}
	resp, httpDoErr := h.httpClient.Do(httpReq)
	if httpDoErr != nil {
		return "", httpDoErr
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image fetch failed: %d", resp.StatusCode)
	}
	data, readImageErr := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if readImageErr != nil || len(data) == 0 {
		return "", fmt.Errorf("empty image")
	}

	ext := filepath.Ext(strings.ToLower(imageURL))
	if ext == "" || len(ext) > 6 {
		ext = ".jpg"
	}
	objectName := fmt.Sprintf("cars/%d_%s_%s%s", time.Now().UnixNano(), strings.ToLower(strings.TrimSpace(make)), strings.ToLower(strings.TrimSpace(model)), ext)
	objectName = strings.ReplaceAll(objectName, " ", "_")
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	return h.storage.Upload(ctx, h.cfg.MinioPublicEndpoint, h.cfg.MinioUseSSL, objectName, bytes.NewReader(data), int64(len(data)), contentType)
}
