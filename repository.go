package multi_tier_caching

import (
	"context"
	"time"
)

// CacheLayer — interface for different cache levels (hot, warm, cold)
type CacheLayer interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string)
	HealthChecker
}

// Database — interface for working with the database
type Database interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

type HealthChecker interface {
	CheckHealth(ctx context.Context) error
}
