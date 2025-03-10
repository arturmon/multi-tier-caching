package storage

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

var (
	redisMetricsOnce sync.Once
	redisMetrics     *RedisMetrics
)

func (r *RedisStorage) initRedisMetrics(client *redis.Client) {
	r.metrics = initRedisMetrics()

	// Launch a goroutine to update metrics
	go r.updateRedisMetrics(client)
}

func (r *RedisStorage) updateRedisMetrics(client *redis.Client) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Retrieve pool stats from the Redis client
		poolStats := client.PoolStats()

		// Update metrics
		r.metrics.PoolHits.Set(float64(poolStats.Hits))
		r.metrics.PoolMisses.Set(float64(poolStats.Misses))
		r.metrics.PoolTimeouts.Set(float64(poolStats.Timeouts))
		r.metrics.PoolTotalConns.Set(float64(poolStats.TotalConns))
		r.metrics.PoolIdleConns.Set(float64(poolStats.IdleConns))
		r.metrics.PoolStaleConns.Set(float64(poolStats.StaleConns))
	}
}

func initRedisMetrics() *RedisMetrics {
	redisMetricsOnce.Do(func() {
		redisMetrics = &RedisMetrics{
			Hits: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "redis_hits_total",
					Help: "Total number of cache hits",
				},
				[]string{"backend"},
			),
			Misses: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "redis_misses_total",
					Help: "Total number of cache misses",
				},
			),
			Writes: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "redis_writes_total",
					Help: "Total number of cache writes",
				},
			),
			PoolHits: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_pool_hits_total",
					Help: "Total number of pool hits",
				},
			),
			PoolMisses: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_pool_misses_total",
					Help: "Total number of pool misses",
				},
			),
			PoolTimeouts: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_pool_timeouts_total",
					Help: "Total number of pool timeouts",
				},
			),
			PoolTotalConns: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_pool_total_connections",
					Help: "Total number of pool connections",
				},
			),
			PoolIdleConns: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_pool_idle_connections",
					Help: "Number of idle pool connections",
				},
			),
			PoolStaleConns: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "redis_pool_stale_connections",
					Help: "Number of stale pool connections",
				},
			),
		}

		// Register all metrics with Prometheus
		prometheus.MustRegister(
			redisMetrics.Hits,
			redisMetrics.Misses,
			redisMetrics.Writes,
			redisMetrics.PoolHits,
			redisMetrics.PoolMisses,
			redisMetrics.PoolTimeouts,
			redisMetrics.PoolTotalConns,
			redisMetrics.PoolIdleConns,
			redisMetrics.PoolStaleConns,
		)
	})
	return redisMetrics
}
