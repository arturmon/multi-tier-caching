package cachelib

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss Add this error definition
var ErrCacheMiss = errors.New("cache miss")

type MultiTierCache struct {
	layers      []CacheLayer
	db          Database
	bloomFilter *BloomFilter
	writeQueue  *WriteQueue
	analytics   *CacheAnalytics
	ttlManager  *TTLManager
}

func NewMultiTierCache(layers []CacheLayer, db Database) *MultiTierCache {
	return &MultiTierCache{
		layers:      layers,
		db:          db,
		bloomFilter: NewBloomFilter(1000000, 5),
		writeQueue:  NewWriteQueue(func(task WriteTask) { _ = db.Set(context.Background(), task.Key, task.Value, 0) }),
		analytics:   NewCacheAnalytics(),
		ttlManager:  NewTTLManager(),
	}
}

func (c *MultiTierCache) Get(ctx context.Context, key string) (string, error) {
	for _, layer := range c.layers {
		value, err := layer.Get(ctx, key)
		if err == nil {
			c.analytics.LogHit()
			return value, nil
		}
	}

	if !c.bloomFilter.Exists(key) {
		c.analytics.LogMiss()
		return "", ErrCacheMiss
	}

	value, err := c.db.Get(ctx, key)
	if err != nil {
		return "", err
	}

	err = c.Set(ctx, key, value)
	if err != nil {
		return "", err
	}
	c.bloomFilter.Add(key)
	return value, nil
}

func (c *MultiTierCache) Set(ctx context.Context, key, value string) error {
	ttl := c.ttlManager.AdjustTTL(key)
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
