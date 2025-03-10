package multi_tier_caching

import (
	"context"
	"testing"

	"github.com/arturmon/multi-tier-caching/mocks"
	databaseMock "github.com/arturmon/multi-tier-caching/storage/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMultiTierCache_Get_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockCache := new(mocks.MockCacheLayer)
	mockDB := new(databaseMock.MockDatabaseStorage)

	// Настраиваем моки
	mockCache.On("Get", ctx, "key1").Return("value1", nil)
	mockCache.On("CheckHealth", ctx).Return(nil)
	mockDB.On("Get", ctx, "key1").Return("value1", nil)
	mockDB.On("CheckHealth", ctx).Return(nil)

	cacheConfig := MultiTierCacheConfig{
		Layers: []LayerInfo{
			{
				Layer: mockCache, // Implements CacheLayer
				Name:  "memory",
			},
			{
				Layer: mockCache, // Implements CacheLayer
				Name:  "redis",
			},
		},
		DB:          mockDB,
		Thresholds:  []int{10, 5},
		BloomSize:   100000,
		BloomHashes: 5,
		Debug:       false,
	}

	cache := NewMultiTierCache(ctx, cacheConfig)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
	mockCache.AssertExpectations(t)
}
