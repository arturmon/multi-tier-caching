# Multi-Tier Cache Library

## Overview
Multi-Tier Cache Library is a modular caching system that provides a multi-layered approach to caching, supporting in-memory, Redis, and database storage. It includes:
- Multi-tier caching
- Bloom filter for cache efficiency
- Write-behind queue for asynchronous updates
- TTL management
- Cache analytics (hit/miss statistics)

## Installation

```sh
go get github.com/arturmon/multi-tier-caching
```

## Usage

### Initialize Cache
```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arturmon/multi-tier-caching"
	"github.com/arturmon/multi-tier-caching-example/config"
	"github.com/arturmon/multi-tier-caching-example/logger"
	"github.com/arturmon/multi-tier-caching/storage"
)


func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.LoadConfig()

	var debugCache = false
	
	logger.InitLogger(cfg.LogLevel)
	logger.Info("Launching the cache system", "memoryCacheSize", cfg.MemoryCacheSize)

	memoryStorage, err := storage.NewRistrettoCache(ctx, int64(cfg.MemoryCacheSize))
	if err != nil {
		logger.Error("Failed to create Memory stoage: %v", err)
	}

	dbStorage, err := storage.NewDatabaseStorage(cfg.DatabaseDSN, debugCache)
	if err != nil {
		logger.Error("Failed to connect to the database", "error", err)
		return
	}
	defer dbStorage.Close()

	redisStorage, err := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword, 1)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return
	}

	logger.Info("Launching the cache system", "memoryCacheSize", cfg.MemoryCacheSize)
	logger.Info("Launching the cache system", "databaseDSN", cfg.DatabaseDSN)
	logger.Info("Launching the cache system", "redisAddr", cfg.RedisAddr)

	cacheConfig := multi_tier_caching.MultiTierCacheConfig{
		Layers: []multi_tier_caching.LayerInfo{
			multi_tier_caching.NewLayerInfo(multi_tier_caching.NewMemoryCache(memoryStorage)),
			multi_tier_caching.NewLayerInfo(multi_tier_caching.NewRedisCache(redisStorage)),
		},
		DB:          multi_tier_caching.NewDatabaseCache(dbStorage),
		Thresholds:  []int{10, 5},
		BloomSize:   100000,
		BloomHashes: 5,
		Debug:       debugCache,
	}

	cache := multi_tier_caching.NewMultiTierCache(ctx, cacheConfig)
	
	err = cache.Set(context.Background(), "key1", "value1")
	if err != nil {
		return
	}
	val, _ := cache.Get(context.Background(), "key1")
	fmt.Println("Cached Value:", val)

	// Waiting for the program to complete
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down...")
}

```

### Using Cache Directly
```go
cache := cachelib.NewMultiTierCache([]cachelib.CacheLayer{
cachelib.NewMemoryCache(),
}, nil)

cache.Set(context.Background(), "key2", "Hello, World!")
value, _ := cache.Get(context.Background(), "key2")
fmt.Println("Cached Value:", value)
```

## Features

- **Multi-tier caching architecture**:
    - Configurable cache layers (e.g., hot, warm, cold) for optimized data access.
    - Prioritizes checks from the hottest layer (in-memory) to colder layers (e.g., disk, remote).

- **Bloom filter optimization**:
    - Reduces unnecessary database queries by probabilistically checking key existence.
    - Dynamically resizes based on request frequency, miss rate, and load factors.

- **Adaptive TTL management**:
    - Adjusts time-to-live (TTL) dynamically using key request frequency.
    - Longer TTL for high-frequency keys to minimize cache churn.

- **Intelligent data migration**:
    - Automatically promotes/demotes keys between layers using frequency thresholds.
    - Processes migrations asynchronously with adjustable intervals via background workers.

- **Metrics and analytics**:
    - Tracks cache hits, misses, migration times, and key frequency.
    - Exposes Prometheus metrics (`cacheHits`, `migrationTime`, `keyFrequency`).

- **Write-behind queue**:
    - Batches and asynchronously persists updates to reduce database latency.
    - Adaptive processing intervals based on queue load.

- **Self-optimizing components**:
    - **Bloom filter auto-scaling**: Dynamically adjusts size and hash functions.
    - **TTL auto-tuning**: Balances cache efficiency and storage costs.

- **Health monitoring**:
    - Health checks for all cache layers and the underlying database.
    - Graceful shutdown with resource cleanup.

- **Debugging and observability**:
    - Detailed logs for migrations, TTL adjustments, and cache operations.
    - `GetDebugInfo()` method for real-time Bloom filter stats (false positive rate, load factor).

- **Configurable policies**:
    - Customizable frequency thresholds for layer transitions.
    - Adjustable Bloom filter parameters (initial size, hash functions).

- **Database integration**:
    - Fallback to database on cache misses, with automatic cache refresh for fetched keys.

- **Concurrency and scalability**:
    - Thread-safe operations using mutexes and channels.
    - Parallel migration workers for high-throughput scenarios.

- **Prometheus integration**:
    - Built-in support for Prometheus metrics (histograms, counters, gauges).


## Contributing
Pull requests are welcome. Please open an issue first to discuss any major changes.

## License
MIT License

