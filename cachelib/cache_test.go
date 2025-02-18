package cachelib

import (
	"context"
	"testing"

	"github.com/arturmon/multi-tier-caching/cachelib/mocks"
	databaseMock "github.com/arturmon/multi-tier-caching/storage/mocks"
	"github.com/stretchr/testify/assert"
)

func TestMultiTierCache_Get_CacheHit(t *testing.T) {
	ctx := context.Background()
	mockCache := new(mocks.MockCacheLayer)
	mockDB := new(databaseMock.MockDatabaseStorage)
	cache := NewMultiTierCache([]CacheLayer{mockCache}, mockDB)

	mockCache.On("Get", ctx, "key1").Return("value1", nil)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
	mockCache.AssertExpectations(t)
}
