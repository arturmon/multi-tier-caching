package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCacheLayer struct {
	mock.Mock
}

func (m *MockCacheLayer) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheLayer) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheLayer) Delete(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *MockCacheLayer) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
