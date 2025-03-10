package multi_tier_caching

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type TTLManager struct {
	ttls  map[string]int64
	mu    sync.Mutex
	debug bool
}

func NewTTLManager(debug bool) *TTLManager {
	prometheus.MustRegister(ttlChangeHistogram)
	return &TTLManager{ttls: make(map[string]int64), debug: debug}
}

func (tm *TTLManager) AdjustTTL(key string, newTTL int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if newTTL > tm.ttls[key] {
		tm.ttls[key] = newTTL
		ttlChangeHistogram.WithLabelValues(key).Observe(float64(newTTL))
	}
}

func (tm *TTLManager) GetTTL(key string) int64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return tm.ttls[key]
}

func (tm *TTLManager) calculateAdaptiveTTL(freq int) int64 {
	baseTTL := int64(60) // Base TTL in seconds
	var ttl int64
	// The higher the frequency, the longer the TTL
	switch {
	case freq > 10:
		ttl = baseTTL * 15 // 15 min
	case freq > 5:
		ttl = baseTTL * 30 // 30 min
	case freq > 2:
		ttl = baseTTL * 60 // 1 hour
	default:
		ttl = baseTTL * 240 // 4 hours
	}
	if tm.debug {
		log.Printf("[TTL] Adaptive TTL calculation: freq=%d → ttl=%d", freq, ttl)
	}
	return ttl
}
