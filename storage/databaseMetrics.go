package storage

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	postgresMetricsOnce sync.Once
	postgresMetrics     *PostgresMetrics
)

func (d *DatabaseStorage) initDatabaseMetrics(pool *pgxpool.Pool) {
	d.metrics = initPostgresMetrics()

	// Launch a goroutine to update metrics
	go d.updateDatabaseMetrics(pool)
}

func (d *DatabaseStorage) updateDatabaseMetrics(pool *pgxpool.Pool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := pool.Stat()
		d.metrics.IdleConns.Set(float64(stats.IdleConns()))
		d.metrics.MaxConnections.Set(float64(stats.MaxConns()))
		d.metrics.AcquiredConns.Set(float64(stats.AcquiredConns()))
		d.metrics.TotalConns.Set(float64(stats.TotalConns()))
	}

}

func initPostgresMetrics() *PostgresMetrics {
	postgresMetricsOnce.Do(func() {
		postgresMetrics = &PostgresMetrics{
			Hits: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "postgres_hits_total",
					Help: "Total number of cache hits",
				},
				[]string{"backend"},
			),
			Misses: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "postgres_misses_total",
					Help: "Total number of cache misses",
				},
			),
			Writes: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "postgres_writes_total",
					Help: "Total number of cache writes",
				},
			),
			QueryCount: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "postgres_query_count_total",
					Help: "Total number of queries executed",
				},
			),
			QueryDuration: prometheus.NewHistogram(
				prometheus.HistogramOpts{
					Name:    "postgres_query_duration_seconds",
					Help:    "Query execution time distribution",
					Buckets: prometheus.DefBuckets,
				},
			),
			MaxConnections: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "postgres_max_connections",
				Help: "Maximum number of PostgreSQL connections",
			}),
			IdleConns: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "postgres_idle_connections",
				Help: "Current idle PostgreSQL connections",
			}),
			AcquiredConns: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "postgres_acquired_connections",
				Help: "Current acquired PostgreSQL connections",
			}),
			TotalConns: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "postgres_total_connections",
				Help: "Total PostgreSQL connections",
			}),
		}

		prometheus.MustRegister(
			postgresMetrics.Hits,
			postgresMetrics.Misses,
			postgresMetrics.Writes,
			postgresMetrics.MaxConnections,
			postgresMetrics.IdleConns,
			postgresMetrics.AcquiredConns,
			postgresMetrics.TotalConns,
		)
	})
	return postgresMetrics
}
