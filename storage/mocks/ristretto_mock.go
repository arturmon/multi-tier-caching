package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockRistrettoStorage — мок для базы данных.
type MockRistrettoStorage struct {
	mock.Mock
}

func (m *MockRistrettoStorage) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRistrettoStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}
func (m *MockRistrettoStorage) CheckHealth(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
