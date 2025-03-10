package multi_tier_caching

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type CacheAnalytics struct {
	cacheHits      *prometheus.CounterVec
	cacheMisses    prometheus.Counter
	keyFrequency   *prometheus.GaugeVec
	migrationTime  prometheus.Histogram
	migrationCount *prometheus.CounterVec
	mu             sync.RWMutex
	keys           map[string]struct{}
}

// LogHit records a cache hit for a specific layer and key, updating the corresponding metrics.
func (a *CacheAnalytics) LogHit(layerName, key string) {
	a.cacheHits.WithLabelValues(layerName).Inc()
	a.keyFrequency.WithLabelValues(key).Inc()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.keys[key] = struct{}{}
}

// LogMiss increments the cache miss counter.
func (a *CacheAnalytics) LogMiss() {
	a.cacheMisses.Inc()
}

// GetFrequency retrieves the request frequency of a specific cache key.
func (a *CacheAnalytics) GetFrequency(key string) int {
	gauge, err := a.keyFrequency.GetMetricWithLabelValues(key)
	if err != nil {
		return 0
	}

	var metric dto.Metric
	if err = gauge.Write(&metric); err != nil {
		return 0
	}

	return int(metric.Gauge.GetValue())
}

// GetFrequencyPerMinute returns a map containing the request frequency of all tracked keys.
func (a *CacheAnalytics) GetFrequencyPerMinute() map[string]int {
	result := make(map[string]int)

	// Create a copy of the keys for secure iteration
	a.mu.RLock()
	keysCopy := make([]string, 0, len(a.keys))
	for key := range a.keys {
		keysCopy = append(keysCopy, key)
	}
	a.mu.RUnlock()

	for _, key := range keysCopy {
		if gauge, err := a.keyFrequency.GetMetricWithLabelValues(key); err == nil {
			var metric dto.Metric
			if err = gauge.Write(&metric); err == nil {
				result[key] = int(*metric.Gauge.Value)
			}
		}
	}
	return result
}
