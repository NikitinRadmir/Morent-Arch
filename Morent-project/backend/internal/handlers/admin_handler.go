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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"morent-backend/internal/config"
	"morent-backend/internal/models"
	adminapp "morent-backend/internal/modules/admin/app"
	admindto "morent-backend/internal/modules/admin/httpdto"
	"morent-backend/internal/modules/transport/http/common"
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
	baseURL := strings.TrimRight(cfg.AggregatorBaseURL, "/")
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

func (h *AdminHandler) Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListUsers(w, r)
	case http.MethodPut:
		h.UpdateUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AdminHandler) Comments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListComments(w, r)
	case http.MethodPut:
		h.UpdateComment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	users, err := h.adminService.ListUsers()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_users",
		Result: "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		common.WriteInternalError(w, "failed to list users")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req admindto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ID == 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err := h.adminService.UpdateUser(req)
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		UserID: req.ID,
		Action: "update_user",
		Data:   req,
		Result: "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		common.WriteInternalError(w, "failed to update user")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	user, _ := h.adminService.GetUser(req.ID)
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err := h.adminService.DeleteUser(id)
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		UserID: id,
		Action: "delete_user",
		Result: "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		common.WriteInternalError(w, "operation failed")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Rentals ----

func (h *AdminHandler) ListRentals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rents, err := h.adminService.ListRentals()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_rentals",
		Result: "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		common.WriteInternalError(w, "operation failed")
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
	err := h.adminService.DeleteRental(uint(id))
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		UserID:   nil,
		ObjectID: id,
		Action:   "delete_rental",
		Result:   "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		common.WriteInternalError(w, "operation failed")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Favorites ----

func (h *AdminHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	favs, err := h.adminService.ListFavorites()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_favorites",
		Result: "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		common.WriteInternalError(w, "operation failed")
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
	err := h.adminService.DeleteFavorite(uint(userID), uint(carID))
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		UserID:   userID,
		ObjectID: carID,
		Action:   "delete_favorite",
		Result:   "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		common.WriteInternalError(w, "operation failed")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Comments ----

func (h *AdminHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	comments, err := h.adminService.ListComments()
	logEvent := service.LogEvent{
		Time:   time.Now(),
		Type:   service.LogCRUD,
		Action: "list_comments",
		Result: "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		common.WriteInternalError(w, "failed to list comments")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	json.NewEncoder(w).Encode(comments)
}

func (h *AdminHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req admindto.UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.ID == 0 {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err := h.adminService.UpdateComment(req)
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		ObjectID: req.ID,
		Action:   "update_comment",
		Data:     req,
		Result:   "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		common.WriteInternalError(w, "failed to update comment")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDFromPath(r.URL.Path)
	if !ok {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	err := h.adminService.DeleteComment(uint(id))
	logEvent := service.LogEvent{
		Time:     time.Now(),
		Type:     service.LogCRUD,
		ObjectID: id,
		Action:   "delete_comment",
		Result:   "success",
	}
	if err != nil {
		logEvent.Result = "error"
		logEvent.Message = err.Error()
		_ = h.logService.LogEvent(ctx, logEvent)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		common.WriteInternalError(w, "operation failed")
		return
	}
	_ = h.logService.LogEvent(ctx, logEvent)
	w.WriteHeader(http.StatusNoContent)
}

// ListLogs возвращает список событий аудита за сегодняшний день.
// Используется в админке для просмотра истории изменений.
func (h *AdminHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	events, err := h.adminService.ListLogs(ctx, time.Now())
	if err != nil {
		common.WriteInternalError(w, "failed to load logs")
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
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		common.WriteBadRequest(w, "Введите поисковый запрос", "query_required")
		return
	}

	body, err := json.Marshal(map[string]string{"q": query})
	if err != nil {
		common.WriteInternalErrorJSON(w, "Не удалось подготовить запрос к агрегатору")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.aggregatorURL+"/search/trims", bytes.NewReader(body))
	if err != nil {
		common.WriteInternalErrorJSON(w, "Не удалось подготовить запрос к агрегатору")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := h.doAggregatorRequest(req)
	if err != nil {
		common.WriteAPIError(w, common.APIError{Error: "Сервис агрегатора временно недоступен", Code: "aggregator_unavailable", Status: http.StatusServiceUnavailable})
		return
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		common.WriteAPIError(w, common.APIError{Error: "Агрегатор вернул некорректный ответ", Code: "aggregator_bad_response", Status: http.StatusBadGateway})
		return
	}

	if resp.StatusCode != http.StatusOK {
		common.WriteAPIError(w, common.APIError{Error: "Агрегатор не смог выполнить поиск", Code: "aggregator_bad_response", Status: http.StatusBadGateway})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBody)
}

func (h *AdminHandler) ImportAggregatorCar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}

	var req ImportAggregatorCarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteBadRequest(w, "Некорректный формат запроса", "invalid_body")
		return
	}
	if strings.TrimSpace(req.Trim.Model) == "" {
		common.WriteBadRequest(w, "В ответе агрегатора отсутствует модель автомобиля", "aggregator_model_required")
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
		if uploadedURL, err := h.uploadAggregatorImage(r.Context(), req.Trim.ImageURL, req.Trim.Make, req.Trim.Model); err == nil && uploadedURL != "" {
			car.ImgSrc = uploadedURL
		}
	}

	if err := h.carService.Create(&car); err != nil {
		common.WriteInternalErrorJSON(w, "Не удалось импортировать автомобиль")
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
	id, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (h *AdminHandler) doAggregatorRequest(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var errRead error
		bodyBytes, errRead = io.ReadAll(req.Body)
		if errRead != nil {
			return nil, errRead
		}
		_ = req.Body.Close()
	}
	getBody := func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(bodyBytes)), nil
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	req.GetBody = getBody
	req.ContentLength = int64(len(bodyBytes))

	var lastErr error
	for attempt := 1; attempt <= h.retryCount; attempt++ {
		cloned := req.Clone(req.Context())
		cloned.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		cloned.GetBody = getBody
		cloned.ContentLength = int64(len(bodyBytes))
		resp, err := h.httpClient.Do(cloned)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		sleepMs := int(math.Min(float64(200*attempt), 1200))
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}
	return nil, lastErr
}

func (h *AdminHandler) uploadAggregatorImage(ctx context.Context, imageURL string, make string, model string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*,*/*;q=0.8")
	req.Header.Set("User-Agent", "Morent/1.0 (+local-dev)")
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("image fetch failed: %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil || len(data) == 0 {
		return "", fmt.Errorf("empty image")
	}

	contentType := strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))
	detectedType := strings.ToLower(http.DetectContentType(data))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectedType
	}
	if !strings.HasPrefix(contentType, "image/") && strings.HasPrefix(detectedType, "image/") {
		contentType = detectedType
	}
	if !isAllowedRemoteImageType(contentType) {
		return "", fmt.Errorf("remote URL is not a supported image: %s", contentType)
	}

	ext := imageExtension(imageURL, contentType)
	objectName := fmt.Sprintf("cars/%d_%s_%s%s", time.Now().UnixNano(), sanitizeObjectPart(make), sanitizeObjectPart(model), ext)
	return h.storage.Upload(ctx, h.cfg.MinioPublicEndpoint, h.cfg.MinioUseSSL, objectName, bytes.NewReader(data), int64(len(data)), contentType)
}

func imageExtension(imageURL string, contentType string) string {
	cleanURL := strings.Split(imageURL, "?")[0]
	ext := strings.ToLower(filepath.Ext(cleanURL))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif":
		return ext
	}
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/avif":
		return ".avif"
	default:
		return ".jpg"
	}
}

func isAllowedRemoteImageType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/jpeg", "image/png", "image/webp", "image/gif", "image/avif":
		return true
	default:
		return false
	}
}

func sanitizeObjectPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}

	var builder strings.Builder
	lastSeparator := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastSeparator = false
			continue
		}
		if !lastSeparator {
			builder.WriteByte('_')
			lastSeparator = true
		}
	}

	cleaned := strings.Trim(builder.String(), "_")
	if cleaned == "" {
		return "unknown"
	}
	return cleaned
}
