package storage

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dgraph-io/ristretto"
)

var ErrCacheMiss = errors.New("cache miss")

// RistrettoCache implements CacheLayer interface using Ristretto
type RistrettoCache struct {
	client  *ristretto.Cache
	metrics *RistrettoMetrics
}

const (
	Heuristic  = 10 * 1000 // Example: 5GB cache limit - 5000
	buffersize = 64
)

// NewRistrettoCache initializes a new Ristretto cache
func NewRistrettoCache(ctx context.Context, memoryLimitMB int64) (*RistrettoCache, error) {
	maxCost := int64(memoryLimitMB * 1024 * 1024)

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: Heuristic,
		MaxCost:     maxCost,
		BufferItems: buffersize,
		Metrics:     true,
	})
	if err != nil {
		log.Fatal("Failed to create Ristretto cache", "error", err)
		return nil, err
	}

	hotStorage := &RistrettoCache{client: cache}
	hotStorage.initRistrettoMetrics(cache)

	// Run background cache cleaning
	go hotStorage.startCacheCleanup(ctx, cache)

	return hotStorage, nil
}

// Get retrieves a value from Ristretto
func (r *RistrettoCache) Get(ctx context.Context, key string) (string, error) {
	value, found := r.client.Get(key)
	if !found {
		r.metrics.Misses.Inc() // metric
		return "", ErrCacheMiss
	}
	r.metrics.Hits.WithLabelValues("ristretto").Inc() // metric
	return value.(string), nil
}

// Set stores a value in Ristretto with TTL
func (r *RistrettoCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	cost := int64(len(value))
	//cost := int64(1)
	ok := r.client.SetWithTTL(key, value, cost, ttl)
	if !ok {
		log.Printf("Error: Failed to write key=%s to Ristretto", key)
		return errors.New("set failed")
	}
	r.client.Wait()        // Ensures that the entry is taken into account in the cache before completion.
	r.metrics.Writes.Inc() // metric
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

// startCacheCleanup Performs periodic cache cleaning.
func (r *RistrettoCache) startCacheCleanup(ctx context.Context, cache *ristretto.Cache) {
	ticker := time.NewTicker(1 * time.Minute) // Cleaning interval
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cache.Clear()
			//log.Println("Cache clearing completed")
		case <-ctx.Done():
			//log.Println("Stop background cache cleaning")
			return
		}
	}
}
