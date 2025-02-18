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
