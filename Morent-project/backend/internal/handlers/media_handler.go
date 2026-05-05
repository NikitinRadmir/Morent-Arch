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

	if parseMultipartFormErr := r.ParseMultipartForm(maxSize); parseMultipartFormErr != nil {
		http.Error(w, "invalid multipart form: "+parseMultipartFormErr.Error(), http.StatusBadRequest)
		return
	}

	file, header, formFileErr := r.FormFile("file")
	if formFileErr != nil {
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
		data, readFileErr := io.ReadAll(file)
		if readFileErr != nil {
			http.Error(w, "failed to read file: "+readFileErr.Error(), http.StatusInternalServerError)
			return
		}
		size = int64(len(data))
		reader = strings.NewReader(string(data))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url, uploadErr := h.storage.Upload(ctx, h.cfg.MinioPublicEndpoint, h.cfg.MinioUseSSL, name, reader, size, header.Header.Get("Content-Type"))
	if uploadErr != nil {
		http.Error(w, "upload failed: "+uploadErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"url":"%s"}`, url)
}
