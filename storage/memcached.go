package storage

import (
	"context"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// MemcachedCache implements CacheLayer interface using Memcached
type MemcachedCache struct {
	client *memcache.Client
}

// NewMemcachedCache initializes a new Memcached client
func NewMemcachedCache(serverAddr string) *MemcachedCache {
	return &MemcachedCache{
		client: memcache.New(serverAddr),
	}
}

// Get retrieves a value from Memcached
func (m *MemcachedCache) Get(ctx context.Context, key string) (string, error) {
	item, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return "", err
		}
		return "", err
	}
	return string(item.Value), nil
}

// Set stores a value in Memcached with TTL
func (m *MemcachedCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return m.client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: int32(ttl.Seconds()),
	})
}

// Delete removes a key from Memcached
func (m *MemcachedCache) Delete(ctx context.Context, key string) {
	m.client.Delete(key)
}
