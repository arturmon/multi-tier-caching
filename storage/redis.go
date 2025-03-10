package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client  *redis.Client
	metrics *RedisMetrics
}

func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// check connect to Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	warmStorage := &RedisStorage{client: client}
	warmStorage.initRedisMetrics(client)
	return warmStorage, nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		r.metrics.Misses.Inc() // metric
		return "", ErrCacheMiss
	} else if err != nil {
		return "", err
	}
	r.metrics.Hits.WithLabelValues("redis").Inc() // metric
	return value, nil
}

func (r *RedisStorage) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	r.client.Set(ctx, key, value, ttl)
	r.metrics.Writes.Inc() // metric
}

func (r *RedisStorage) Delete(ctx context.Context, key string) {
	r.client.Del(ctx, key)
}

func (r *RedisStorage) CheckHealth(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
