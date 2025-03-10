package multi_tier_caching

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"
)

// ErrCacheMiss Add this error definition
var ErrCacheMiss = errors.New("cache miss")

type MultiTierCache struct {
	layers         []LayerInfo // Cache layers sorted from hot to cold
	db             Database
	bloomFilter    *BloomFilter
	writeQueue     *WriteQueue
	analytics      *CacheAnalytics
	migration      *MigrationManager
	ttlManager     *TTLManager
	freqThresholds []int // Request rate thresholds for transition between layers
	debug          bool  //
}
type MultiTierCacheConfig struct {
	Layers      []LayerInfo // Cache layers sorted from hot to cold
	DB          Database
	Thresholds  []int
	BloomSize   uint
	BloomHashes uint
	Debug       bool
}

func NewMultiTierCache(ctx context.Context, config MultiTierCacheConfig) *MultiTierCache {
	if len(config.Thresholds) != len(config.Layers) {
		panic("The number of thresholds (thresholds) must be equal to the number of cache layers (layers)")
	}
	var layersInfo []LayerInfo
	for _, layer := range config.Layers {
		layersInfo = append(layersInfo, LayerInfo{
			Layer: layer.Layer,
			Name:  layer.Layer.String(),
		})
	}
	ttlManager := NewTTLManager(config.Debug)
	analytics := NewCacheAnalytics()

	bloomFilter := NewBloomFilter(config.BloomSize, config.BloomHashes, config.Debug, analytics)

	migrationMgr := NewMigrationManager(
		layersInfo,
		ttlManager,
		analytics,
		config.Thresholds,
		config.DB,
		config.Debug,
	)

	cache := &MultiTierCache{
		layers:      layersInfo,
		db:          config.DB,
		bloomFilter: bloomFilter,
		writeQueue: NewWriteQueue(func(task WriteTask) {
			_ = config.DB.Set(ctx, task.Key, task.Value, task.TTL)
		}, config.Debug),
		analytics:  analytics,
		migration:  migrationMgr,
		ttlManager: ttlManager,
		debug:      config.Debug,
	}

	// Background process for migrating data between layers
	migrationMgr.Start(ctx)
	return cache
}

func NewLayerInfo(layer CacheLayer) LayerInfo {
	return LayerInfo{
		Layer: layer,
		Name:  layer.String(),
	}
}

func (c *MultiTierCache) Get(ctx context.Context, key string) (string, error) {
	// First we look in the hottest layer (the first layer)
	value, err := c.layers[0].Layer.Get(ctx, key)
	if err == nil {
		c.analytics.LogHit(fmt.Sprintf("layer_%s", c.layers[0].Name), key)
		if c.debug {
			log.Printf("[CACHE] Found key=%s in hot layer", key)
		}
		return value, nil
	}

	// If not found in the hot layer, continue searching in other layers
	for i := 1; i < len(c.layers); i++ {
		value, err = c.layers[i].Layer.Get(ctx, key)
		if err == nil {
			c.analytics.LogHit(fmt.Sprintf("layer_%s", c.layers[i].Name), key)
			if c.debug {
				log.Printf("[CACHE] Found key=%s in layer=%d (%v)", key, i, c.layers[i].Name)
			}
			return value, nil
		}
	}

	// If not found in the layer and the Bloom filter does not exclude the key
	if !c.bloomFilter.Exists(key) {
		c.analytics.LogMiss()
		return "", ErrCacheMiss
	}

	// If you didn't find it in the cache, go to the database
	value, err = c.db.Get(ctx, key)
	if err != nil {
		c.analytics.LogMiss()
		return "", err
	}
	c.analytics.LogHit("database", key)

	// Refreshing cache and Bloom filter
	freq := c.analytics.GetFrequency(key)
	targetLayers := c.selectTargetLayers(freq)
	for _, layer := range targetLayers {
		if c.debug {
			log.Printf("[CACHE] Key=%s frequency: %d target layers: %v", key, freq, layer.Name)
		}
	}

	// Refresh the cache in the selected layers
	err = c.initCachePlacement(ctx, key, value, targetLayers)
	if err != nil {
		return "", err
	}

	c.bloomFilter.Add(key)
	return value, nil
}

func (c *MultiTierCache) Set(ctx context.Context, key, value string) error {
	freq := c.analytics.GetFrequency(key) // We get the frequency of requests
	adaptiveTTL := c.ttlManager.calculateAdaptiveTTL(freq)
	currentTTL := c.ttlManager.GetTTL(key)
	if c.debug {
		log.Printf("[CACHE] Current TTL for key=%s: %d, new adaptive TTL: %d", key, currentTTL, adaptiveTTL)
	}

	// Always update TTL for hot keys
	if c.isKeyInHotLayer(ctx, key) {
		adaptiveTTL = c.ttlManager.calculateAdaptiveTTL(freq)
		currentTTL = adaptiveTTL // Forced update
	}
	// Set TTL only if it is greater than the current one
	if int64(adaptiveTTL) > currentTTL {
		ttlSeconds := time.Duration(adaptiveTTL) * time.Second
		targetLayers := c.selectTargetLayers(freq)
		for _, layerInfo := range targetLayers {
			if err := layerInfo.Layer.Set(ctx, key, value, ttlSeconds); err != nil {
				log.Printf("Error writing to layer: %v", err)
				return err
			}
		}
		c.ttlManager.AdjustTTL(key, int64(adaptiveTTL))
		c.writeQueue.Enqueue(WriteTask{Key: key, Value: value, TTL: ttlSeconds})
		c.bloomFilter.Add(key)
	}
	return nil
}

func (c *MultiTierCache) HealthCheck(ctx context.Context) error {
	// Checking all cache layers
	for _, layer := range c.layers {
		if err := layer.Layer.CheckHealth(ctx); err != nil {
			return fmt.Errorf("cache layer error: %w", err)
		}
	}

	// Database check
	if db, ok := c.db.(HealthChecker); ok {
		if err := db.CheckHealth(ctx); err != nil {
			return fmt.Errorf("database error: %w", err)
		}
	}

	return nil
}

func (c *MultiTierCache) Close() {
	if closer, ok := c.db.(io.Closer); ok {
		closer.Close()
	}
	c.writeQueue.Stop()
}

func (c *MultiTierCache) selectTargetLayers(freq int) []LayerInfo {
	var layers []LayerInfo
	for i, threshold := range c.freqThresholds {
		if freq >= threshold {
			layers = append(layers, LayerInfo{
				Layer: c.layers[i].Layer,
				Name:  c.layers[i].Layer.String(),
			})
		}
	}
	return layers
}

func (c *MultiTierCache) initCachePlacement(ctx context.Context, key, value string, targetLayers []LayerInfo) error {
	// Get the current TTL of the key
	currentTTL := c.ttlManager.GetTTL(key)
	// Calculating a new frequency-based adaptive TTL
	freq := c.analytics.GetFrequency(key)
	adaptiveTTL := int64(c.ttlManager.calculateAdaptiveTTL(freq))

	// We update TTL only if the new one is greater than the current one
	if adaptiveTTL > currentTTL {
		c.ttlManager.AdjustTTL(key, adaptiveTTL)
		currentTTL = adaptiveTTL
	}

	ttlSeconds := time.Duration(currentTTL) * time.Second

	// Update target layers with current TTL
	for _, layerInfo := range targetLayers {
		if err := layerInfo.Layer.Set(ctx, key, value, ttlSeconds); err != nil {
			log.Printf("[CACHE] Error writing to layer %v: %v", layerInfo.Name, err)
			return err
		}
		if c.debug {
			log.Printf("[CACHE] Successfully recorded in %v: key=%s, TTL=%v", layerInfo.Name, key, ttlSeconds)
		}
	}

	c.bloomFilter.Add(key)
	// Add a task to the queue with the current TTL
	c.writeQueue.Enqueue(WriteTask{Key: key, Value: value, TTL: ttlSeconds})
	return nil
}

// Checking for the presence of a key in the hot layer
func (c *MultiTierCache) isKeyInHotLayer(ctx context.Context, key string) bool {
	_, err := c.layers[0].Layer.Get(ctx, key)
	return err == nil
}
