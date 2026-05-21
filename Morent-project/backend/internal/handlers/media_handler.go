package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"morent-backend/internal/config"
	"morent-backend/internal/storage"
)

var allowedUploadMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

type MediaHandler struct {
	cfg     *config.Config
	storage *storage.MinioStorage
}

func NewMediaHandler(cfg *config.Config, storage *storage.MinioStorage) *MediaHandler {
	return &MediaHandler{cfg: cfg, storage: storage}
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const maxSize = 20 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxSize))
	if err != nil {
		http.Error(w, "failed to read file", http.StatusBadRequest)
		return
	}
	if len(data) == 0 {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}

	sniffLen := len(data)
	if sniffLen > 512 {
		sniffLen = 512
	}
	contentType := http.DetectContentType(data[:sniffLen])
	if !allowedUploadMIME[contentType] {
		http.Error(w, "unsupported file type", http.StatusBadRequest)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch contentType {
	case "image/jpeg":
		if ext != ".jpg" && ext != ".jpeg" {
			ext = ".jpg"
		}
	case "image/png":
		if ext != ".png" {
			ext = ".png"
		}
	case "image/webp":
		if ext != ".webp" {
			ext = ".webp"
		}
	case "image/gif":
		if ext != ".gif" {
			ext = ".gif"
		}
	}
	name := uuid.NewString() + ext

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	url, err := h.storage.Upload(ctx, h.cfg.MinioPublicEndpoint, h.cfg.MinioUseSSL, name, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}
