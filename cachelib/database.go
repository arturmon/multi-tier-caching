package cachelib

import (
	"context"
	"time"

	"github.com/arturmon/multi-tier-caching/storage"
)

type DatabaseCache struct {
	storage *storage.DatabaseStorage
}

// NewDatabaseCache initializes a new database cache
func NewDatabaseCache(db *storage.DatabaseStorage) *DatabaseCache {
	return &DatabaseCache{storage: db}
}

// Get retrieves the value from the database cache
func (d *DatabaseCache) Get(ctx context.Context, key string) (string, error) {
	value, err := d.storage.GetCache(ctx, key)
	if err != nil {
		return "", err
	}
	return value, err
}

// Set stores a value in the database cache with TTL
func (d *DatabaseCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := d.storage.SetCache(ctx, key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a value from the database cache
func (d *DatabaseCache) Delete(ctx context.Context, key string) {
	err := d.storage.DeleteCache(ctx, key)
	if err != nil {
		return
	}
}
