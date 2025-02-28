package storage

import (
	"context"
	"errors"
	"time"

	"github.com/dgraph-io/ristretto"
)

var ErrCacheMiss = errors.New("cache miss")

// RistrettoCache implements CacheLayer interface using Ristretto
type RistrettoCache struct {
	client *ristretto.Cache
}

// NewRistrettoCache initializes a new Ristretto cache
func NewRistrettoCache(sizeKB int64) (*RistrettoCache, error) {
	sizeBytes := sizeKB * 1024
	avgObjectSize := int64(100)
	maxCost := sizeBytes / avgObjectSize
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxCost * 10, // Number of counters for TinyLFU (approx. 10x MaxCost)
		MaxCost:     maxCost,      // Maximum memory consumption (notional cost)
		BufferItems: 64,           // Number of buffer records for concurrent operations
	})
	if err != nil {
		return nil, err
	}
	return &RistrettoCache{client: cache}, nil
}

// Get retrieves a value from Ristretto
func (r *RistrettoCache) Get(ctx context.Context, key string) (string, error) {
	value, found := r.client.Get(key)
	if !found {
		return "", ErrCacheMiss
	}
	return value.(string), nil
}

// Set stores a value in Ristretto with TTL
func (r *RistrettoCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	r.client.SetWithTTL(key, value, int64(len(value)), ttl)
	r.client.Wait() // Ensures that the entry is taken into account in the cache before completion.
	return nil
}

// Delete removes a key from Ristretto
func (r *RistrettoCache) Delete(ctx context.Context, key string) {
	r.client.Del(key)
}

func (r *RistrettoCache) CheckHealth(ctx context.Context) error {
	// Ristretto has no connection state, we only check for initialization
	if r.client == nil {
		return errors.New("ristretto cache is not initialized")
	}
	return nil
}
