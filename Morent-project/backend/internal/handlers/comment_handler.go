package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"morent-backend/internal/config"
	"morent-backend/internal/models"
	commentsdto "morent-backend/internal/modules/comments/httpdto"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/service"
)

type CommentHandler struct {
	service     *service.CommentService
	authService *service.AuthService
	cfg         *config.Config
}

var commentValidator = validator.New()

func NewCommentHandler(service *service.CommentService, authService *service.AuthService, cfg *config.Config) *CommentHandler {
	return &CommentHandler{service: service, authService: authService, cfg: cfg}
}

func (h *CommentHandler) authenticate(r *http.Request) (*models.User, error) {
	return common.Authenticate(h.authService, h.cfg, r)
}

func (h *CommentHandler) GetCarComments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		common.WriteBadRequest(w, "Некорректный путь запроса", "invalid_path")
		return
	}

	carIDStr := parts[len(parts)-1]
	carID, errParse := strconv.ParseUint(carIDStr, 10, 32)
	if errParse != nil {
		common.WriteBadRequest(w, "Некорректный идентификатор автомобиля", "invalid_car_id")
		return
	}

	comments, errGet := h.service.GetByCarID(int(carID))
	if errGet != nil {
		common.WriteInternalErrorJSON(w, "Не удалось загрузить комментарии")
		return
	}

	jsonData, errEncode := json.Marshal(comments)
	if errEncode != nil {
		common.WriteInternalErrorJSON(w, "Не удалось сформировать ответ")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}

	user, err := h.authenticate(r)
	if err != nil || user == nil {
		common.WriteUnauthorized(w, "Войдите в аккаунт")
		return
	}

	var payload commentsdto.CreateCommentRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&payload); errDecode != nil {
		common.WriteBadRequest(w, "Некорректный формат запроса", "invalid_body")
		return
	}
	if err := commentValidator.Struct(payload); err != nil {
		common.WriteValidationError(w, err)
		return
	}

	canComment, errCan := h.service.UserCanComment(user.ID, payload.CarID)
	if errCan != nil {
		common.WriteInternalErrorJSON(w, "Не удалось проверить право на комментарий")
		return
	}
	if !canComment {
		common.WriteForbidden(w, "Оставить отзыв можно только после аренды этого автомобиля")
		return
	}

	comment := models.Comment{
		CarID:       payload.CarID,
		UserID:      user.ID,
		Name:        user.Nickname,
		Post:        user.Position,
		Photo:       user.AvatarURL,
		Date:        "just now",
		Rating:      payload.Rating,
		Description: strings.TrimSpace(payload.Description),
		CreatedAt:   time.Now(),
	}

	if errCreate := h.service.Create(&comment); errCreate != nil {
		common.WriteInternalErrorJSON(w, "Не удалось сохранить комментарий")
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}

	user, errAuth := h.authenticate(r)
	if errAuth != nil || user == nil {
		common.WriteUnauthorized(w, "Войдите в аккаунт")
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		common.WriteBadRequest(w, "Некорректный путь запроса", "invalid_path")
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		common.WriteBadRequest(w, "Некорректный идентификатор комментария", "invalid_comment_id")
		return
	}

	existing, errGet := h.service.GetByID(uint(id))
	if errGet != nil {
		common.WriteInternalError(w, "failed to load comment")
		return
	}
	if existing == nil {
		common.WriteNotFound(w, "Комментарий не найден")
		return
	}
	if existing.UserID != user.ID && strings.ToLower(user.Role) != "admin" {
		common.WriteForbidden(w, "Недостаточно прав для изменения комментария")
		return
	}

	var payload commentsdto.UpdateCommentBody
	if errDecode := json.NewDecoder(r.Body).Decode(&payload); errDecode != nil {
		common.WriteBadRequest(w, "Некорректный формат запроса", "invalid_body")
		return
	}
	comment := *existing
	if payload.Description != nil {
		comment.Description = strings.TrimSpace(*payload.Description)
	}
	if payload.Rating != nil {
		comment.Rating = *payload.Rating
	}

	if errUpdate := h.service.Update(&comment); errUpdate != nil {
		if errors.Is(errUpdate, gorm.ErrRecordNotFound) {
			common.WriteNotFound(w, "Комментарий не найден")
			return
		}
		common.WriteInternalError(w, "failed to update comment")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		common.WriteAPIError(w, common.APIError{Error: "method not allowed", Code: "method_not_allowed", Status: http.StatusMethodNotAllowed})
		return
	}

	user, errAuth := h.authenticate(r)
	if errAuth != nil || user == nil {
		common.WriteUnauthorized(w, "Войдите в аккаунт")
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		common.WriteBadRequest(w, "Некорректный путь запроса", "invalid_path")
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		common.WriteBadRequest(w, "Некорректный идентификатор комментария", "invalid_comment_id")
		return
	}

	existing, errGet := h.service.GetByID(uint(id))
	if errGet != nil {
		common.WriteInternalError(w, "failed to load comment")
		return
	}
	if existing == nil {
		common.WriteNotFound(w, "Комментарий не найден")
		return
	}
	if existing.UserID != user.ID && strings.ToLower(user.Role) != "admin" {
		common.WriteForbidden(w, "Недостаточно прав для удаления комментария")
		return
	}

	if errDelete := h.service.Delete(uint(id)); errDelete != nil {
		if errors.Is(errDelete, gorm.ErrRecordNotFound) {
			common.WriteNotFound(w, "Комментарий не найден")
			return
		}
		common.WriteInternalError(w, "failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
