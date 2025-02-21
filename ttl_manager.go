package multi_tier_caching

import "sync"

type TTLManager struct {
	ttls map[string]int64
	mu   sync.Mutex
}

func NewTTLManager() *TTLManager {
	return &TTLManager{ttls: make(map[string]int64)}
}

func (tm *TTLManager) AdjustTTL(key string) int64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	ttl := tm.ttls[key] + 10
	tm.ttls[key] = ttl
	return ttl
}

func (c *MultiTierCache) calculateAdaptiveTTL(freq int) int {
	baseTTL := 60 // Base TTL in seconds

	// The higher the frequency, the longer the TTL
	switch {
	case freq > 100:
		return baseTTL * 10 // 10 min
	case freq > 50:
		return baseTTL * 5 // 5 min
	case freq > 20:
		return baseTTL * 2 // 2 min
	default:
		return baseTTL // 1 min
	}
}
