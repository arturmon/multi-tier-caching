package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockDatabaseStorage — мок для базы данных.
type MockDatabaseStorage struct {
	mock.Mock
}

func (m *MockDatabaseStorage) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockDatabaseStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}
func (m *MockDatabaseStorage) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
