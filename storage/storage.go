package storage

import "github.com/prometheus/client_golang/prometheus"

/*
type Debuggable interface {
	GetDebugInfo() map[string]interface{}
}

*/

type CacheMetrics struct {
	PostgresMetrics
	RedisMetrics
	RistrettoMetrics
}

// PostgresMetrics Unique PostgreSQL Metrics
type PostgresMetrics struct {
	Hits           *prometheus.CounterVec
	Misses         prometheus.Counter
	Writes         prometheus.Counter
	QueryCount     prometheus.Counter
	QueryDuration  prometheus.Histogram
	MaxConnections prometheus.Gauge
	IdleConns      prometheus.Gauge
	AcquiredConns  prometheus.Gauge
	TotalConns     prometheus.Gauge
}

// RedisMetrics Unique Redis Metrics
type RedisMetrics struct {
	Hits           *prometheus.CounterVec
	Misses         prometheus.Counter
	Writes         prometheus.Counter
	PoolHits       prometheus.Gauge
	PoolMisses     prometheus.Gauge
	PoolTimeouts   prometheus.Gauge
	PoolTotalConns prometheus.Gauge
	PoolIdleConns  prometheus.Gauge
	PoolStaleConns prometheus.Gauge
}

// RistrettoMetrics Unique Ristretto Metrics
type RistrettoMetrics struct {
	Hits            *prometheus.CounterVec
	Misses          prometheus.Counter
	Writes          prometheus.Counter
	CostAdded       prometheus.Counter
	CostEvicted     prometheus.Counter
	KeysAdded       prometheus.Counter
	KeysEvicted     prometheus.Counter
	Entries         prometheus.Gauge   // For tracking the number of entries in the cache
	CacheHits       prometheus.Counter // For tracking total cache hits
	CacheMisses     prometheus.Counter // For tracking total cache misses
	ClientCostAdded prometheus.Counter // For tracking client cost added
	CacheWrites     prometheus.Gauge
}
