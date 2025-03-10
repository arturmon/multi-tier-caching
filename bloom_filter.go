package multi_tier_caching

import (
	"log"
	"sync"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/prometheus/client_golang/prometheus"
)

type BloomFilter struct {
	debug          bool
	filter         *bloom.BloomFilter
	mu             sync.Mutex
	analytics      *CacheAnalytics
	increaseFactor float64
	decreaseFactor float64
	maxSize        uint
	minSize        uint
	lastAdjustment time.Time
	stopChan       chan struct{}
}

func NewBloomFilter(size uint, hashFuncs uint, debug bool, analytics *CacheAnalytics) *BloomFilter {
	prometheus.MustRegister(
		bloomCapacityGauge,
		bloomCountGauge,
		bloomHashFunctionsGauge,
		bloomFalsePositiveGauge,
		bloomLoadFactorGauge,
		bloomLastAdjustmentGauge,
	)
	bloomFilter := &BloomFilter{
		filter:         bloom.New(size, hashFuncs),
		debug:          debug,
		analytics:      analytics,
		increaseFactor: 1.5,
		decreaseFactor: 0.7,
		maxSize:        100000,
		minSize:        1000,
		lastAdjustment: time.Now(),
		stopChan:       make(chan struct{}),
	}
	go bloomFilter.metricsUpdater()
	return bloomFilter
}

func (b *BloomFilter) Add(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.filter.AddString(key)
	if b.debug {
		log.Printf("[BLOOM] Added key: %s", key)
	}
	// Checking the load and changing the filter size
	b.adjustFilterSize()
}

func (b *BloomFilter) Exists(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	exists := b.filter.TestString(key)
	if b.debug {
		log.Printf("[BLOOM] Check key=%s, exists=%v", key, exists)
	}
	return exists
}

func (b *BloomFilter) adjustFilterSize() {
	if time.Since(b.lastAdjustment) < 20*time.Second {
		if b.debug {
			log.Printf("[BLOOM] Skipping adjustment, lastAdjustment was %v", b.lastAdjustment)
		}
		return // Adjust maximum once per minute
	}

	currentSize := b.filter.Cap()
	metrics := b.analytics.GetFrequencyPerMinute()
	totalRequests := 0
	for _, v := range metrics {
		totalRequests += v
	}

	// Check cache miss rate
	misses := getCounterValue(b.analytics.cacheMisses)
	hits := getCounterValueVec(b.analytics.cacheHits)

	// Calculate desired size based on metrics
	desiredSize := uint(float64(currentSize) * b.calculateAdjustmentFactor(totalRequests, misses, hits))

	// Apply boundaries
	if desiredSize > b.maxSize {
		desiredSize = b.maxSize
	} else if desiredSize < b.minSize {
		desiredSize = b.minSize
	}

	if desiredSize != currentSize {
		b.resizeFilter(desiredSize, metrics)
		b.lastAdjustment = time.Now()
		if b.debug {
			log.Printf("[BLOOM] Resized filter to %d, lastAdjustment updated to %s", desiredSize, b.lastAdjustment.Format("15:04:05"))
		}
	}
}

func (b *BloomFilter) calculateAdjustmentFactor(totalRequests int, misses float64, hits float64) float64 {
	if totalRequests == 0 {
		return 1.0
	}

	missRate := misses / (hits + misses)
	loadFactor := float64(len(b.analytics.keys)) / float64(b.filter.Cap())

	// Increase size if high miss rate or load factor > 0.75
	if missRate > 0.1 || loadFactor > 0.75 {
		return b.increaseFactor
	}

	// Decrease size if low utilization and miss rate
	if loadFactor < 0.25 && missRate < 0.05 {
		return b.decreaseFactor
	}

	return 1.0
}

func (b *BloomFilter) resizeFilter(newSize uint, frequencies map[string]int) {
	newFilter := bloom.New(newSize, b.filter.K())

	// Dynamic threshold based on average number of requests
	total := 0
	for _, count := range frequencies {
		total += count
	}
	avg := total / len(frequencies)
	threshold := avg // or another coefficient, for example, avg / 2

	for key, count := range frequencies {
		if count > threshold {
			newFilter.AddString(key)
		}
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.filter = newFilter

	if b.debug {
		log.Printf("[BLOOM] Resized filter from %d to %d, kept %d keys",
			b.filter.Cap(), newSize, len(frequencies))
	}
}

func (b *BloomFilter) metricsUpdater() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//b.adjustFilterSize()
			b.UpdateMetrics()
		case <-b.stopChan:
			return
		}
	}
}

func (b *BloomFilter) UpdateMetrics() {
	b.mu.Lock()
	defer b.mu.Unlock()

	capacity := float64(b.filter.Cap())
	count := float64(b.filter.ApproximatedSize())
	hashFuncs := float64(b.filter.K())
	falsePositiveRate := estimateFalsePositiveRate(b.filter.K(), b.filter.Cap(), uint(count))
	loadFactor := count / capacity
	lastAdjustmentTS := float64(b.lastAdjustment.Unix())

	bloomCapacityGauge.Set(capacity)
	bloomCountGauge.Set(count)
	bloomHashFunctionsGauge.Set(hashFuncs)
	bloomFalsePositiveGauge.Set(falsePositiveRate)
	bloomLoadFactorGauge.Set(loadFactor)
	bloomLastAdjustmentGauge.Set(lastAdjustmentTS)

	if b.debug {
		log.Printf("[BLOOM] Metrics updated - "+
			"Capacity: %.0f, Count: %.0f, Hashes: %.0f, "+
			"FPR: %.6f, Load: %.4f, LastAdj: %d",
			capacity, count, hashFuncs,
			falsePositiveRate, loadFactor, int64(lastAdjustmentTS))
	}
}
