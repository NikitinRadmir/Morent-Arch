package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"morent-backend/internal/storage"
)

type LogService struct {
	storage *storage.MinioStorage
}

func NewLogService(storage *storage.MinioStorage) *LogService {
	return &LogService{storage: storage}
}

type LogEventType string

const (
	LogCRUD     LogEventType = "CRUD"
	LogAuth     LogEventType = "AUTH"
	LogProfile  LogEventType = "PROFILE"
	LogFavorite LogEventType = "FAVORITE"
	LogRental   LogEventType = "RENTAL"
)

type LogEvent struct {
	Time      time.Time     `json:"time"`
	Type      LogEventType  `json:"type"`
	UserID    interface{}   `json:"userId,omitempty"`
	ObjectID  interface{}   `json:"objectId,omitempty"`
	Action    string        `json:"action"`
	Data      interface{}   `json:"data,omitempty"`
	Result    string        `json:"result"`
	Message   string        `json:"message,omitempty"`
}

func (s *LogService) LogEvent(ctx context.Context, event LogEvent) error {
	payload, marshalPayloadErr := json.Marshal(event)
	if marshalPayloadErr != nil {
		return marshalPayloadErr
	}
	return s.storage.AppendDailyLog(ctx, string(payload))
}

// GetDailyEvents читает события из лога MinIO за указанный день.
// Используется для отображения истории изменений в админке.
func (s *LogService) GetDailyEvents(ctx context.Context, day time.Time) ([]LogEvent, error) {
	objectName := fmt.Sprintf("logs/%s.log", day.Format("2006-01-02"))
	obj, getObjectErr := s.storage.Client.GetObject(ctx, s.storage.Bucket, objectName, minio.GetObjectOptions{})
	if getObjectErr != nil {
		return nil, getObjectErr
	}
	defer obj.Close()

	data, readAllErr := io.ReadAll(obj)
	if readAllErr != nil {
		return nil, readAllErr
	}

	lines := strings.Split(string(data), "\n")
	events := make([]LogEvent, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Каждая строка: "<timestamp> {json}"
		idx := strings.Index(line, "{")
		if idx == -1 {
			continue
		}
		jsonPart := line[idx:]
		var ev LogEvent
		if unmarshalErr := json.Unmarshal([]byte(jsonPart), &ev); unmarshalErr != nil {
			continue
		}
		events = append(events, ev)
	}
	return events, nil
}


