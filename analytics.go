package multi_tier_caching

import (
	"sync"
	"time"
)

type CacheAnalytics struct {
	hits, misses  int
	frequencies   map[string]int // Request frequency for each keyword
	recentFreq    sync.Map       // Request frequency in the last minute
	mu            sync.Mutex
	lastResetTime time.Time
}

func NewCacheAnalytics() *CacheAnalytics {
	return &CacheAnalytics{
		frequencies:   make(map[string]int),
		lastResetTime: time.Now(),
	}
}

func (ca *CacheAnalytics) LogHit(key string) {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	ca.hits++
	ca.frequencies[key]++
	valRecent, _ := ca.recentFreq.LoadOrStore(key, 0)
	ca.recentFreq.Store(key, valRecent.(int)+1)
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

func (ca *CacheAnalytics) GetFrequency(key string) int {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return ca.frequencies[key]
}

// GetTotalFrequency Returns the total number of times a key has been accessed.
func (ca *CacheAnalytics) GetTotalFrequency(key string) int {
	ca.mu.Lock()
	defer ca.mu.Unlock()
	return ca.frequencies[key]
}

// GetFrequencyPerMinute Returns the request rate per minute and resets the counters
func (ca *CacheAnalytics) GetFrequencyPerMinute() map[string]int {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	now := time.Now()
	if now.Sub(ca.lastResetTime).Minutes() < 1 {
		return nil // Not a minute has passed yet
	}

	freqCopy := make(map[string]int)
	// Iterating over the sync.Map using Range
	ca.recentFreq.Range(func(k, v interface{}) bool {
		if key, ok := k.(string); ok {
			if count, ok := v.(int); ok {
				freqCopy[key] = count
			}
		}
		return true // continue iteration
	})

	// Resetting temporary data
	ca.recentFreq = sync.Map{}
	ca.lastResetTime = now

	return freqCopy
}

// ResetFrequencies We reset the counters every minute
func (ca *CacheAnalytics) ResetFrequencies() {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	now := time.Now()
	if now.Sub(ca.lastResetTime).Minutes() >= 1 {
		ca.frequencies = make(map[string]int) // Clearing the counters
		ca.lastResetTime = now
	}
}
