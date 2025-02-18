package multi_tier_caching

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheAnalytics(t *testing.T) {
	analytics := NewCacheAnalytics()

	// Checking the initial values
	hits, misses := analytics.GetStats()
	assert.Equal(t, 0, hits, "The initial number of hits should be 0")
	assert.Equal(t, 0, misses, "The initial number of misses should be 0")

	// We log 3 hits
	analytics.LogHit()
	analytics.LogHit()
	analytics.LogHit()

	// Logging 2 misses
	analytics.LogMiss()
	analytics.LogMiss()

	// Checking the updated values
	hits, misses = analytics.GetStats()
	assert.Equal(t, 3, hits, "The number of hits should be 3")
	assert.Equal(t, 2, misses, "The number of misses should be 2")
}
