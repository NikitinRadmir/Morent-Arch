package cache

import (
	"context"
	"fmt"
	"time"

	"morent-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis подключает Redis; при пустом адресе возвращает nil без ошибки.
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	if cfg.RedisAddr == "" {
		return nil, nil
	}
	opts := &redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rdb, nil
}
