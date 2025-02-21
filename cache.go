package multi_tier_caching

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss Add this error definition
var ErrCacheMiss = errors.New("cache miss")

type MultiTierCache struct {
	layers         []CacheLayer // Cache layers sorted from hot to cold
	db             Database
	bloomFilter    *BloomFilter
	writeQueue     *WriteQueue
	analytics      *CacheAnalytics
	ttlManager     *TTLManager
	freqThresholds []int // Request rate thresholds for transition between layers
}

func NewMultiTierCache(layers []CacheLayer, db Database, thresholds []int) *MultiTierCache {
	cache := &MultiTierCache{
		layers:         layers,
		db:             db,
		bloomFilter:    NewBloomFilter(1000000, 5),
		writeQueue:     NewWriteQueue(func(task WriteTask) { _ = db.Set(context.Background(), task.Key, task.Value, 0) }),
		analytics:      NewCacheAnalytics(),
		ttlManager:     NewTTLManager(),
		freqThresholds: thresholds,
	}

	// Starting a background process to clean frequencies
	go func() {
		for {
			time.Sleep(60 * time.Second)
			cache.analytics.ResetFrequencies()
		}
	}()

	return cache
}

func (c *MultiTierCache) Get(ctx context.Context, key string) (string, error) {
	// Searching in cache layers
	for _, layer := range c.layers {
		value, err := layer.Get(ctx, key)
		if err == nil {
			c.analytics.LogHit(key) // Logging a hit with a key
			return value, nil
		}
	}
	// Checking in Bloom filter
	if !c.bloomFilter.Exists(key) {
		c.analytics.LogMiss()
		return "", ErrCacheMiss
	}
	// Getting from DB
	value, err := c.db.Get(ctx, key)
	if err != nil {
		return "", err
	}

	// Defining target layers based on frequency
	freq := c.analytics.GetFrequency(key)
	targetLayers := c.selectTargetLayers(freq)

	// Update cache only in selected layers
	err = c.updateLayers(ctx, key, value, targetLayers)
	if err != nil {
		return "", err
	}

	c.bloomFilter.Add(key)
	return value, nil
}

func (c *MultiTierCache) Set(ctx context.Context, key, value string) error {
	freq := c.analytics.GetFrequency(key)
	ttl := c.calculateAdaptiveTTL(freq) // New Method for Adaptive TTL
	ttlSeconds := time.Duration(ttl) * time.Second

	for _, layer := range c.layers {
		err := layer.Set(ctx, key, value, ttlSeconds)
		if err != nil {
			return err
		}
	}

	c.writeQueue.Enqueue(WriteTask{Key: key, Value: value})
	return nil
}

func (c *MultiTierCache) selectTargetLayers(freq int) []CacheLayer {
	var layers []CacheLayer

	// Adaptive movement: frequently requested data stays in fast layers
	for i := len(c.freqThresholds) - 1; i >= 0; i-- {
		if freq >= c.freqThresholds[i] {
			layers = append(layers, c.layers[i])
		}
	}
	return layers
}

func (c *MultiTierCache) updateLayers(ctx context.Context, key, value string, targetLayers []CacheLayer) error {
	ttl := c.ttlManager.AdjustTTL(key)
	ttlSeconds := time.Duration(ttl) * time.Second

	// We update only target layers
	for _, layer := range targetLayers {
		if err := layer.Set(ctx, key, value, ttlSeconds); err != nil {
			return err
		}
	}

	// Clearing the key from layers not included in the target
	for _, layer := range c.layers {
		if !containsLayer(targetLayers, layer) {
			layer.Delete(ctx, key)
		}
	}

	c.writeQueue.Enqueue(WriteTask{Key: key, Value: value})
	return nil
}

func containsLayer(layers []CacheLayer, target CacheLayer) bool {
	for _, l := range layers {
		if l == target {
			return true
		}
	}
	return false
}
