package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"morent-backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	Client *minio.Client
	Bucket string
}

func NewMinioStorage(cfg *config.Config) (*MinioStorage, error) {
	client, initErr := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if initErr != nil {
		return nil, fmt.Errorf("minio init failed: %w", initErr)
	}

	ctx := context.Background()
	exists, bucketExistsErr := client.BucketExists(ctx, cfg.MinioBucket)
	if bucketExistsErr != nil {
		return nil, fmt.Errorf("minio check bucket: %w", bucketExistsErr)
	}
	if !exists {
		if createBucketErr := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); createBucketErr != nil {
			return nil, fmt.Errorf("minio create bucket: %w", createBucketErr)
		}
	}

	publicPolicy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, cfg.MinioBucket)
	_ = client.SetBucketPolicy(ctx, cfg.MinioBucket, publicPolicy)

	return &MinioStorage{
		Client: client,
		Bucket: cfg.MinioBucket,
	}, nil
}

func (s *MinioStorage) Upload(ctx context.Context, endpoint string, useSSL bool, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	opts := minio.PutObjectOptions{ContentType: contentType}
	if _, putObjectErr := s.Client.PutObject(ctx, s.Bucket, objectName, reader, size, opts); putObjectErr != nil {
		return "", fmt.Errorf("minio put object: %w", putObjectErr)
	}
	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint, s.Bucket, objectName)
	return url, nil
}

// AppendDailyLog appends a log line to a daily log file in MinIO (logs/YYYY-MM-DD.log).
// If the file does not exist, it will be created. Content is rewritten on each append.
func (s *MinioStorage) AppendDailyLog(ctx context.Context, message string) error {
	today := time.Now().Format("2006-01-02")
	objectName := fmt.Sprintf("logs/%s.log", today)
	line := fmt.Sprintf("%s %s\n", time.Now().Format(time.RFC3339), message)

	// Try to read existing content (if any).
	var existing []byte
	obj, getObjectErr := s.Client.GetObject(ctx, s.Bucket, objectName, minio.GetObjectOptions{})
	if getObjectErr == nil {
		existing, _ = io.ReadAll(obj) // ignore read errors, treat as empty
		_ = obj.Close()
	}

	newContent := append(existing, []byte(line)...)
	reader := bytes.NewReader(newContent)

	if _, putLogErr := s.Client.PutObject(ctx, s.Bucket, objectName, reader, int64(len(newContent)), minio.PutObjectOptions{ContentType: "text/plain"}); putLogErr != nil {
		return fmt.Errorf("minio append log: %w", putLogErr)
	}
	return nil
}