package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseStorage struct {
	debug   bool
	pool    *pgxpool.Pool
	metrics *PostgresMetrics
}

// NewDatabaseStorage initializes a connection to PostgreSQL via pgx/v5
func NewDatabaseStorage(dsn string, debug bool) (*DatabaseStorage, error) {
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

	// Create a function to clean expired cache entries
	_, err = pool.Exec(ctx, `
		CREATE OR REPLACE FUNCTION clean_expired_cache() 
		RETURNS TRIGGER AS $$
		BEGIN
			DELETE FROM cache WHERE expires_at <= NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create clean_expired_cache function: %w", err)
	}

	// Create a trigger to call the function after insert or update
	_, err = pool.Exec(ctx, `
				DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT 1 FROM pg_trigger WHERE tgname = 'clean_expired_cache_trigger'
					) THEN
						CREATE TRIGGER clean_expired_cache_trigger
						AFTER INSERT OR UPDATE ON cache
						FOR EACH ROW
						EXECUTE FUNCTION clean_expired_cache();
					END IF;
				END $$;
			`)

	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to create trigger: %w", err)
	}

	dbStorage := &DatabaseStorage{pool: pool, debug: debug}
	dbStorage.initDatabaseMetrics(pool)
	return dbStorage, nil
}

// GetCache gets a value from the cache by key
func (d *DatabaseStorage) GetCache(ctx context.Context, key string) (string, error) {
	if d.debug {
		log.Printf("[DB CACHE] Getting key=%s", key)
	}
	var value string
	start := time.Now()
	err := d.pool.QueryRow(ctx,
		"SELECT value FROM cache WHERE key = $1 AND expires_at > NOW()",
		key,
	).Scan(&value)

	d.metrics.QueryCount.Inc()
	d.metrics.QueryDuration.Observe(time.Since(start).Seconds())

	if errors.Is(err, pgx.ErrNoRows) {
		d.metrics.Misses.Inc() // Increasing misses
		return "", nil
	} else if err != nil {
		return "", err
	}

	d.metrics.Hits.WithLabelValues("PostgreSQL").Inc() // Increasing hits
	return value, nil
}

// SetCache set value in the cache with a TTL
func (d *DatabaseStorage) SetCache(ctx context.Context, key string, value string, ttl time.Duration) error {
	if d.debug {
		log.Printf("[DB CACHE] Setting key=%s with TTL: %v", key, ttl)
	}
	start := time.Now()
	_, err := d.pool.Exec(ctx, `
		INSERT INTO cache (key, value, expires_at) 
		VALUES ($1, $2, NOW() + $3::interval) 
		ON CONFLICT (key) 
		DO UPDATE SET value = $2, expires_at = NOW() + $3::interval`,
		key, value, ttl.String())
	d.metrics.QueryCount.Inc()
	d.metrics.QueryDuration.Observe(time.Since(start).Seconds())
	d.metrics.Writes.Inc() // Increment writes
	return err
}

// DeleteCache removes a value from the cache by key
func (d *DatabaseStorage) DeleteCache(ctx context.Context, key string) error {
	start := time.Now()
	_, err := d.pool.Exec(ctx, "DELETE FROM cache WHERE key = $1", key)
	d.metrics.QueryCount.Inc()
	d.metrics.QueryDuration.Observe(time.Since(start).Seconds())
	return err
}

// Close closes the connection to the database
func (d *DatabaseStorage) Close() {
	d.pool.Close()
}

func (d *DatabaseStorage) CheckHealth(ctx context.Context) error {
	conn, err := d.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("database connection error: %w", err)
	}
	defer conn.Release()
	return conn.Ping(ctx)
}
