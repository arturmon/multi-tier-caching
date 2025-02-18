package cachelib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloomFilter(t *testing.T) {
	filter := NewBloomFilter(1000, 5)

	// Check if the key is not in the filter
	key := "test_key"
	assert.False(t, filter.Exists(key), "The key should not exist in an empty filter")

	// Add the key and check again
	filter.Add(key)
	assert.True(t, filter.Exists(key), "The key should exist after adding")

	// Check for another key that is not there
	assert.False(t, filter.Exists("other_key"), "The other key should not exist in the filter")
}
