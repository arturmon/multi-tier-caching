package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	MemoryCacheSize int
	RedisAddr       string
	RedisPassword   string
	DatabaseDSN     string
	LogLevel        string
}

func LoadConfig() *Config {
	// Load environment variables from .env if it exists
	_ = godotenv.Load()

	memoryCacheSize, err := strconv.Atoi(getEnv("MEMORY_CACHE_SIZE", "1000"))
	if err != nil {
		log.Fatalf("Ошибка при парсинге MEMORY_CACHE_SIZE: %v", err)
	}

	return &Config{
		MemoryCacheSize: memoryCacheSize,
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		DatabaseDSN:     getEnv("DATABASE_DSN", "postgres://user:password@localhost:5432/dbname"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

// Helper function to get value from ENV with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
