package multi_tier_caching

import (
	"context"
	"time"

	"github.com/arturmon/multi-tier-caching/storage"
)

type MemoryCache struct {
	storage *storage.RistrettoCache
}

// NewMemoryCache initializes a new database cache
func NewMemoryCache(ram *storage.RistrettoCache) *MemoryCache {
	return &MemoryCache{storage: ram}
}

// Get retrieves the value from the database cache
func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	value, err := m.storage.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return value, err
}

// Set stores a value in the database cache with TTL
func (m *MemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := m.storage.Set(ctx, key, value, ttl)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a value from the database cache
func (m *MemoryCache) Delete(ctx context.Context, key string) {
	m.storage.Delete(ctx, key)
}

func (m *MemoryCache) CheckHealth(ctx context.Context) error {
	return m.storage.CheckHealth(ctx)
}

func (m *MemoryCache) String() string {
	return "Ristretto"
}
