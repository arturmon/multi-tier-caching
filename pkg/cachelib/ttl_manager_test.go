package cachelib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTTLManager(t *testing.T) {
	tm := NewTTLManager()

	key := "test_key"

	// Check the initial TTL (it should be 10, since initially 0 + 10)
	ttl := tm.AdjustTTL(key)
	assert.Equal(t, int64(10), ttl, "Initial TTL should be 10")

	// Increase TTL several times
	ttl = tm.AdjustTTL(key)
	assert.Equal(t, int64(20), ttl, "TTL should increase to 20")

	ttl = tm.AdjustTTL(key)
	assert.Equal(t, int64(30), ttl, "TTL should increase to 30")
}
