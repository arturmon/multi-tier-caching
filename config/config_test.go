package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Set environment variables for testing
	os.Setenv("MEMORY_CACHE_SIZE", "2048")
	os.Setenv("REDIS_ADDR", "localhost:6380")
	os.Setenv("REDIS_PASSWORD", "testpassword")
	os.Setenv("DATABASE_DSN", "postgres://testuser:testpassword@localhost:5432/testdb")
	os.Setenv("LOG_LEVEL", "debug")

	// Load config
	config := LoadConfig()

	// Test values
	if config.MemoryCacheSize != 2048 {
		t.Errorf("Expected MEMORY_CACHE_SIZE 2048, but got %d", config.MemoryCacheSize)
	}

	if config.RedisAddr != "localhost:6380" {
		t.Errorf("Expected REDIS_ADDR 'localhost:6380', but got %s", config.RedisAddr)
	}

	if config.RedisPassword != "testpassword" {
		t.Errorf("Expected REDIS_PASSWORD 'testpassword', but got %s", config.RedisPassword)
	}

	if config.DatabaseDSN != "postgres://testuser:testpassword@localhost:5432/testdb" {
		t.Errorf("Expected DATABASE_DSN 'postgres://testuser:testpassword@localhost:5432/testdb', but got %s", config.DatabaseDSN)
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected LOG_LEVEL 'debug', but got %s", config.LogLevel)
	}

	// Clean up environment variables after test
	os.Unsetenv("MEMORY_CACHE_SIZE")
	os.Unsetenv("REDIS_ADDR")
	os.Unsetenv("REDIS_PASSWORD")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("LOG_LEVEL")
}
