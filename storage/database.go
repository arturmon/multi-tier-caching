package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseStorage struct {
	pool *pgxpool.Pool
}

// NewDatabaseStorage initializes a connection to PostgreSQL via pgx/v5
func NewDatabaseStorage(dsn string) (*DatabaseStorage, error) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	// Ensure the cache table exists
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS cache (
			key VARCHAR(255) PRIMARY KEY,
			value TEXT,
			expires_at TIMESTAMPTZ
		)
	`)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create cache table: %w", err)
	}

	return &DatabaseStorage{pool: pool}, nil
}

// GetCache gets a value from the cache by key
func (d *DatabaseStorage) GetCache(ctx context.Context, key string) (string, error) {
	var value string
	err := d.pool.QueryRow(ctx, "SELECT value FROM cache WHERE key = $1 AND expires_at > NOW()", key).Scan(&value)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetCache set value in the cache with a TTL
func (d *DatabaseStorage) SetCache(ctx context.Context, key string, value string, ttl time.Duration) error {
	_, err := d.pool.Exec(ctx, `
		INSERT INTO cache (key, value, expires_at) 
		VALUES ($1, $2, NOW() + $3::interval) 
		ON CONFLICT (key) 
		DO UPDATE SET value = $2, expires_at = NOW() + $3::interval`,
		key, value, ttl.String())
	return err
}

// DeleteCache removes a value from the cache by key
func (d *DatabaseStorage) DeleteCache(ctx context.Context, key string) error {
	_, err := d.pool.Exec(ctx, "DELETE FROM cache WHERE key = $1", key)
	return err
}

// Close closes the connection to the database
func (d *DatabaseStorage) Close() {
	d.pool.Close()
}
