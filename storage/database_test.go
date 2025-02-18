package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/arturmon/multi-tier-caching/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDatabaseStorage_GetCache(t *testing.T) {
	// Create a mock of the DatabaseStorage
	mockStorage := new(mocks.MockDatabaseStorage)

	// Define test cases
	tests := []struct {
		key         string
		mockReturn  string
		mockError   error
		expected    string
		expectError bool
	}{
		{"key1", "value1", nil, "value1", false},
		{"key2", "", errors.New("cache not found"), "", true},
	}

	// Run each test case
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			// Set up the mock behavior
			mockStorage.On("Get", mock.Anything, test.key).Return(test.mockReturn, test.mockError)

			// Call the method
			result, err := mockStorage.Get(context.Background(), test.key)

			// Check the result
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}

			// Ensure the mock expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestDatabaseStorage_SetCache(t *testing.T) {
	// Create a mock of the DatabaseStorage
	mockStorage := new(mocks.MockDatabaseStorage)

	// Define test cases
	tests := []struct {
		key         string
		value       string
		ttl         time.Duration
		mockError   error
		expectError bool
	}{
		{"key1", "value1", time.Minute, nil, false},
		{"key2", "value2", time.Hour, errors.New("set cache failed"), true},
	}

	// Run each test case
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			// Set up the mock behavior
			mockStorage.On("Set", mock.Anything, test.key, test.value).Return(test.mockError)

			// Call the method
			err := mockStorage.Set(context.Background(), test.key, test.value, 0)

			// Check the result
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Ensure the mock expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}

/*
func TestDatabaseStorage_DeleteCache(t *testing.T) {
	// Create a mock of the DatabaseStorage
	mockStorage := new(mocks.MockDatabaseStorage)

	// Define test cases
	tests := []struct {
		key         string
		mockError   error
		expectError bool
	}{
		{"key1", nil, false},
		{"key2", errors.New("delete failed"), true},
	}

	// Run each test case
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			// Set up the mock behavior
			mockStorage.On("Delete", mock.Anything, test.key).Return(test.mockError)

			// Call the method
			err := mockStorage.Delete(context.Background(), test.key)

			// Check the result
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Ensure the mock expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}

*/
