package multi_tier_caching

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache_Get_Set(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	// Test cache miss
	value, err := cache.Get(ctx, "missing_key")
	assert.ErrorIs(t, err, ErrCacheMiss)
	assert.Equal(t, "", value)

	// Test setting and getting a key
	err = cache.Set(ctx, "key1", "value1", 5*time.Second)
	assert.NoError(t, err)

	value, err = cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
}

func TestMemoryCache_Expiration(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	cache.Set(ctx, "temp_key", "temp_value", 5*time.Second)

	// Ensure value is retrievable before TTL expiration
	value, err := cache.Get(ctx, "temp_key")
	assert.NoError(t, err)
	assert.Equal(t, "temp_value", value)

	// Wait for the TTL to expire
	time.Sleep(6 * time.Second) // Slightly more than 5 minutes

	value, err = cache.Get(ctx, "temp_key")
	assert.ErrorIs(t, err, ErrCacheMiss)
	assert.Equal(t, "", value)
}

func TestMemoryCache_Delete(t *testing.T) {
	ctx := context.Background()
	cache := NewMemoryCache()

	cache.Set(ctx, "key1", "value1", 5*time.Second)

	// Ensure value exists
	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Delete the key
	cache.Delete(ctx, "key1")

	// Ensure key is removed
	value, err = cache.Get(ctx, "key1")
	assert.ErrorIs(t, err, ErrCacheMiss)
	assert.Equal(t, "", value)
}
