package storage

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ristrettoMetricsOnce sync.Once
	ristrettoMetrics     *RistrettoMetrics
)

func (r *RistrettoCache) initRistrettoMetrics(client *ristretto.Cache) {
	r.metrics = initRistrettoMetrics()

	// Start a goroutine to update metrics
	go r.updateRistrettoMetrics(client)
}

func (r *RistrettoCache) updateRistrettoMetrics(client *ristretto.Cache) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Retrieve pool stats from the Ristretto client
		poolStats := client.Metrics

		// Update metrics
		r.metrics.Hits.WithLabelValues("backend").Add(float64(poolStats.Hits()))
		r.metrics.Misses.Add(float64(poolStats.Misses()))
		r.metrics.CostAdded.Add(float64(poolStats.CostAdded()))
		r.metrics.CostEvicted.Add(float64(poolStats.CostEvicted()))
		r.metrics.Writes.Add(float64(poolStats.KeysAdded()))
		r.metrics.KeysAdded.Add(float64(poolStats.KeysAdded()))
		r.metrics.KeysEvicted.Add(float64(poolStats.KeysEvicted()))
		r.metrics.Entries.Set(float64(poolStats.KeysAdded() - poolStats.KeysEvicted())) // Entries = KeysAdded - KeysEvicted
		r.metrics.ClientCostAdded.Add(float64(r.client.Metrics.CostAdded()))            // Assuming r.client.Metrics has CostAdded field
	}
}

func initRistrettoMetrics() *RistrettoMetrics {
	ristrettoMetricsOnce.Do(func() {
		ristrettoMetrics = &RistrettoMetrics{
			Hits: prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "ristretto_cache_cache_hits_total",
					Help: "Total number of Ristretto cache hits",
				},
				[]string{"backend"},
			),
			Misses: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_cache_cache_misses_total",
					Help: "Total number of Ristretto cache misses",
				},
			),
			Writes: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_cache_cache_writes_total",
					Help: "Total number of Ristretto cache writes",
				},
			),
			CostAdded: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_cache_cost_added_total",
					Help: "Total cost added to Ristretto cache",
				},
			),
			CostEvicted: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_cache_cost_evicted_total",
					Help: "Total cost evicted from Ristretto cache",
				},
			),
			KeysAdded: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_cache_keys_added_total",
					Help: "Total number of keys added to Ristretto cache",
				},
			),
			KeysEvicted: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_keys_evicted_total",
					Help: "Total number of keys evicted from Ristretto cache",
				},
			),
			Entries: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "ristretto_entries",
					Help: "Number of entries in Ristretto cache",
				},
			),
			ClientCostAdded: prometheus.NewCounter(
				prometheus.CounterOpts{
					Name: "ristretto_client_cost_added_total",
					Help: "Total client cost added to Ristretto cache",
				},
			),
			CacheWrites: prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: "ristretto_cache_writes_total",
					Help: "Total number of cache writes",
				},
			),
		}

		// Register all metrics with Prometheus
		prometheus.MustRegister(
			ristrettoMetrics.Hits,
			ristrettoMetrics.Misses,
			ristrettoMetrics.Writes,
			ristrettoMetrics.CostAdded,
			ristrettoMetrics.CostEvicted,
			ristrettoMetrics.KeysAdded,
			ristrettoMetrics.KeysEvicted,
			ristrettoMetrics.Entries,
			ristrettoMetrics.ClientCostAdded,
			ristrettoMetrics.CacheWrites,
		)
	})

	return ristrettoMetrics
}
