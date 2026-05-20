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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	carIDStr := parts[len(parts)-1]
	carID, errParse := strconv.ParseUint(carIDStr, 10, 32)
	if errParse != nil {
		http.Error(w, "Invalid car ID format", http.StatusBadRequest)
		return
	}

	comments, errGet := h.service.GetByCarID(int(carID))
	if errGet != nil {
		http.Error(w, "Error fetching comments: "+errGet.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, errEncode := json.Marshal(comments)
	if errEncode != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.authenticate(r)
	if err != nil || user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload commentsdto.CreateCommentRequest
	if errDecode := json.NewDecoder(r.Body).Decode(&payload); errDecode != nil {
		http.Error(w, "Invalid request body: "+errDecode.Error(), http.StatusBadRequest)
		return
	}
	if err := commentValidator.Struct(payload); err != nil {
		http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	canComment, errCan := h.service.UserCanComment(user.ID, payload.CarID)
	if errCan != nil {
		http.Error(w, "Error checking permission: "+errCan.Error(), http.StatusInternalServerError)
		return
	}
	if !canComment {
		http.Error(w, "User has not rented this car", http.StatusForbidden)
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
		http.Error(w, "Error creating comment: "+errCreate.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, errAuth := h.authenticate(r)
	if errAuth != nil || user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	existing, errGet := h.service.GetByID(uint(id))
	if errGet != nil {
		common.WriteInternalError(w, "failed to load comment")
		return
	}
	if existing == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	if existing.UserID != user.ID && strings.ToLower(user.Role) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var payload commentsdto.UpdateCommentBody
	if errDecode := json.NewDecoder(r.Body).Decode(&payload); errDecode != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error updating comment: "+errUpdate.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, errAuth := h.authenticate(r)
	if errAuth != nil || user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	idStr := parts[len(parts)-1]
	id, errParse := strconv.ParseUint(idStr, 10, 32)
	if errParse != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	existing, errGet := h.service.GetByID(uint(id))
	if errGet != nil {
		common.WriteInternalError(w, "failed to load comment")
		return
	}
	if existing == nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	if existing.UserID != user.ID && strings.ToLower(user.Role) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if errDelete := h.service.Delete(uint(id)); errDelete != nil {
		if errors.Is(errDelete, gorm.ErrRecordNotFound) {
			http.Error(w, "Comment not found", http.StatusNotFound)
			return
		}
		common.WriteInternalError(w, "failed to delete comment")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
