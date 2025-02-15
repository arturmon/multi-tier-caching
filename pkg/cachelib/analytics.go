package cachelib

import (
	"sync"
)

type CacheAnalytics struct {
	hits, misses int
	mu           sync.Mutex
}

func NewCacheAnalytics() *CacheAnalytics {
	return &CacheAnalytics{}
}

func (ca *CacheAnalytics) LogHit() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.hits++
}

func (ca *CacheAnalytics) LogMiss() {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.misses++
}

func (ca *CacheAnalytics) GetStats() (int, int) {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return ca.hits, ca.misses
}
