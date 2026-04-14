package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"morent-backend/internal/config"
	"morent-backend/internal/storage"
)

type MediaHandler struct {
	cfg     *config.Config
	storage *storage.MinioStorage
}

func NewMediaHandler(cfg *config.Config, storage *storage.MinioStorage) *MediaHandler {
	return &MediaHandler{cfg: cfg, storage: storage}
}

// Upload принимает multipart/form-data с полем "file" и загружает его в MinIO.
// Возвращает JSON: { "url": "<public-url>" }.
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ограничим размер запроса (например, 20 МБ).
	const maxSize = 20 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		http.Error(w, "invalid multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	name := uuid.NewString() + ext

	// Определяем размер; если неизвестен, читаем в буфер.
	var reader io.Reader = file
	var size int64 = header.Size
	if size <= 0 {
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		size = int64(len(data))
		reader = strings.NewReader(string(data))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url, err := h.storage.Upload(ctx, h.cfg.MinioEndpoint, h.cfg.MinioUseSSL, name, reader, size, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"url":"%s"}`, url)
}

