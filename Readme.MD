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

	cfg := config.LoadConfig()

	logger.InitLogger(cfg.LogLevel)
	logger.Info("Launching the cache system", "memoryCacheSize", cfg.MemoryCacheSize)

	dbStorage, err := storage.NewDatabaseStorage(cfg.DatabaseDSN)
	if err != nil {
		logger.Error("Failed to connect to the database", "error", err)
		return
	}
	defer dbStorage.Close()

	redisStorage, err := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		return
	}

	cache := multi_tier_caching.NewMultiTierCache(
		[]multi_tier_caching.CacheLayer{
			multi_tier_caching.NewMemoryCache(),
			multi_tier_caching.NewRedisCache(redisStorage),
		},
		multi_tier_caching.NewDatabaseCache(dbStorage),
	)

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
- **Multi-tier caching**: Uses memory, Redis, and database storage layers.
- **Bloom filter**: Reduces unnecessary database queries.
- **Write-behind queue**: Efficient database updates.
- **TTL Management**: Supports automatic expiration of cached items.
- **Cache Analytics**: Provides cache hit/miss statistics.

## Contributing
Pull requests are welcome. Please open an issue first to discuss any major changes.

## License
MIT License

