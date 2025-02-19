package multi_tier_caching

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/arturmon/multi-tier-caching/storage"
)

var (
	// ErrInvalidType indicates the stored value is not a string.
	ErrInvalidType = errors.New("invalid type stored in cache")
)

type MemoryCache struct {
	storage sync.Map
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

// NewCacheLayer returns the appropriate cache layer based on the environment variable
func NewCacheLayer(useMemcached bool, memcachedAddr string) CacheLayer {
	if useMemcached == true {
		return storage.NewMemcachedCache(memcachedAddr)
	}
	return NewMemoryCache()
}

// Get now takes a context and returns a string.
func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	value, exists := m.storage.Load(key)
	if !exists {
		return "", ErrCacheMiss
	}
	str, ok := value.(string)
	if !ok {
		return "", ErrInvalidType
	}
	return str, nil
}

// Set now takes a context as well.
func (m *MemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.storage.Store(key, value)
	// Automatic deletion by TTL
	go func() {
		select {
		case <-ctx.Done():
			// If the context is cancelled, stop the timer
			return
		case <-time.After(ttl):
			m.storage.Delete(key)
		}
	}()
	return nil
}

// Delete now takes a context.
func (m *MemoryCache) Delete(ctx context.Context, key string) {
	m.storage.Delete(key)
}
