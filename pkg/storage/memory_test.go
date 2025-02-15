package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStorage_Set(t *testing.T) {
	// Create a new MemoryStorage instance
	storage := NewMemoryStorage()

	// Test setting and getting a value
	storage.Set("key1", "value1", time.Second)

	// Validate that the value is correctly stored
	value, ok := storage.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// Test that the value expires after the TTL
	time.Sleep(2 * time.Second)

	// After TTL, the key should be deleted
	_, ok = storage.Get("key1")
	assert.False(t, ok)
}

func TestMemoryStorage_Delete(t *testing.T) {
	// Create a new MemoryStorage instance
	storage := NewMemoryStorage()

	// Test setting and deleting a value
	storage.Set("key2", "value2", time.Second)

	// Validate that the value is correctly stored
	value, ok := storage.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)

	// Now delete the key
	storage.Delete("key2")

	// Validate that the key is deleted
	_, ok = storage.Get("key2")
	assert.False(t, ok)
}

func TestMemoryStorage_GetNonExistentKey(t *testing.T) {
	// Create a new MemoryStorage instance
	storage := NewMemoryStorage()

	// Test retrieving a non-existent key
	_, ok := storage.Get("non-existent-key")
	assert.False(t, ok)
}
