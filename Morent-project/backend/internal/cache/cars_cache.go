package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"morent-backend/internal/models"

	"github.com/redis/go-redis/v9"
)

const carsKeyPrefix = "morent:cars:"

// CarCache кэширует выдачу машин в Redis (опционально: при nil-клиенте методы no-op).
type CarCache struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewCarCache(rdb *redis.Client, ttl time.Duration) *CarCache {
	if rdb == nil || ttl <= 0 {
		return &CarCache{}
	}
	return &CarCache{rdb: rdb, ttl: ttl}
}

func (c *CarCache) enabled() bool {
	return c != nil && c.rdb != nil && c.ttl > 0
}

func filterKey(name, carType string, capacity *int, priceUnder *float64) string {
	raw := fmt.Sprintf("%s|%s|", name, carType)
	if capacity != nil {
		raw += strconv.Itoa(*capacity)
	}
	raw += "|"
	if priceUnder != nil {
		raw += fmt.Sprintf("%g", *priceUnder)
	}
	sum := sha256.Sum256([]byte(raw))
	return carsKeyPrefix + "f:" + hex.EncodeToString(sum[:])
}

func (c *CarCache) GetAll(ctx context.Context) ([]models.Car, bool, error) {
	if !c.enabled() {
		return nil, false, nil
	}
	val, err := c.rdb.Get(ctx, carsKeyPrefix+"all").Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var cars []models.Car
	if errUnmarshal := json.Unmarshal(val, &cars); errUnmarshal != nil {
		return nil, false, errUnmarshal
	}
	return cars, true, nil
}

func (c *CarCache) SetAll(ctx context.Context, cars []models.Car) error {
	if !c.enabled() {
		return nil
	}
	b, err := json.Marshal(cars)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, carsKeyPrefix+"all", b, c.ttl).Err()
}

func (c *CarCache) GetByID(ctx context.Context, id int) (*models.Car, bool, error) {
	if !c.enabled() {
		return nil, false, nil
	}
	key := fmt.Sprintf("%sid:%d", carsKeyPrefix, id)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var car models.Car
	if errUnmarshal := json.Unmarshal(val, &car); errUnmarshal != nil {
		return nil, false, errUnmarshal
	}
	return &car, true, nil
}

func (c *CarCache) SetByID(ctx context.Context, id int, car *models.Car) error {
	if !c.enabled() || car == nil {
		return nil
	}
	b, err := json.Marshal(car)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("%sid:%d", carsKeyPrefix, id)
	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

func (c *CarCache) GetFiltered(ctx context.Context, name, carType string, capacity *int, priceUnder *float64) ([]models.Car, bool, error) {
	if !c.enabled() {
		return nil, false, nil
	}
	key := filterKey(name, carType, capacity, priceUnder)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var cars []models.Car
	if errUnmarshal := json.Unmarshal(val, &cars); errUnmarshal != nil {
		return nil, false, errUnmarshal
	}
	return cars, true, nil
}

func (c *CarCache) SetFiltered(ctx context.Context, name, carType string, capacity *int, priceUnder *float64, cars []models.Car) error {
	if !c.enabled() {
		return nil
	}
	b, err := json.Marshal(cars)
	if err != nil {
		return err
	}
	key := filterKey(name, carType, capacity, priceUnder)
	return c.rdb.Set(ctx, key, b, c.ttl).Err()
}

// InvalidateCars удаляет все ключи кэша машин (после мутаций).
func (c *CarCache) InvalidateCars(ctx context.Context) error {
	if !c.enabled() {
		return nil
	}
	var cursor uint64
	for {
		keys, next, errScan := c.rdb.Scan(ctx, cursor, carsKeyPrefix+"*", 100).Result()
		if errScan != nil {
			return errScan
		}
		if len(keys) > 0 {
			if errDel := c.rdb.Del(ctx, keys...).Err(); errDel != nil {
				return errDel
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
