package multi_tier_caching

import "github.com/prometheus/client_golang/prometheus"

var ttlChangeHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "cache_ttl_changes",
		Help:    "Histogram of TTL values assigned to cache keys",
		Buckets: []float64{60, 300, 600, 1800, 3600, 7200, 14400}, // 1m, 5m, 10m, 30m, 1h, 2h, 4h
	},
	[]string{"key"},
)
