package mocks

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// MockRedisStorage — мок для Redis.
type MockRedisStorage struct {
	mock.Mock
}

func (m *MockRedisStorage) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisStorage) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, ttl)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisStorage) Del(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}
