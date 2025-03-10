package multi_tier_caching

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NewCacheAnalytics initializes and returns a new CacheAnalytics instance with Prometheus metrics.
func NewCacheAnalytics() *CacheAnalytics {
	return &CacheAnalytics{
		cacheHits: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total cache hits per layer",
		}, []string{"layer"}),

		cacheMisses: promauto.NewCounter(prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total cache misses",
		}),

		keyFrequency: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cache_key_frequency",
			Help: "Request frequency for keys",
		}, []string{"key"}),

		migrationTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "cache_migration_duration_seconds",
			Help:    "Time spent on data migration",
			Buckets: []float64{0.1, 0.5, 1, 5},
		}),

		migrationCount: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "cache_migration_operations_total",
			Help: "Total migration operations",
		}, []string{"status"}),
		keys: make(map[string]struct{}),
	}
}
