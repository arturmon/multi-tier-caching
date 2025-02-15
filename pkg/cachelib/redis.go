package cachelib

import (
	"context"
	"time"

	"github.com/arturmon/multi-tier-caching/pkg/storage"
)

type RedisCache struct {
	storage *storage.RedisStorage
}

func NewRedisCache(client *storage.RedisStorage) *RedisCache {
	return &RedisCache{storage: client}
}

// Get now takes a context and returns a string.
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	value, err := r.storage.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return value, err
}

// Set now takes a context.
func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	r.storage.Set(ctx, key, value, ttl)
	return nil
}

// Delete now takes a context.
func (r *RedisCache) Delete(ctx context.Context, key string) {
	r.storage.Delete(ctx, key)
}
