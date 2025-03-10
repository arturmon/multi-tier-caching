package multi_tier_caching

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type MigrationManager struct {
	layers         []LayerInfo
	ttlManager     *TTLManager
	analytics      *CacheAnalytics
	migrationQueue chan string
	thresholds     []int
	db             Database
	debug          bool
}

func NewMigrationManager(
	layers []LayerInfo,
	ttlMgr *TTLManager,
	analytics *CacheAnalytics,
	thresholds []int,
	db Database,
	debug bool,
) *MigrationManager {
	return &MigrationManager{
		layers:         layers,
		ttlManager:     ttlMgr,
		analytics:      analytics,
		thresholds:     thresholds,
		migrationQueue: make(chan string, 1000),
		db:             db,
		debug:          debug,
	}
}

func (m *MigrationManager) Start(ctx context.Context) {
	for i := 0; i < 5; i++ { // 5 workers
		go m.scheduleDynamic(ctx)
	}
}

func (m *MigrationManager) scheduleDynamic(ctx context.Context) {
	if m.debug {
		log.Printf("[MIGRATION] Worker started") // Worker startup log
	}
	ticker := time.NewTicker(m.calculateInterval())
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processMigrations(ctx)

		case key := <-m.migrationQueue:
			startTime := time.Now()
			if m.debug {
				log.Printf("[MIGRATION] Processing key from queue: %s", key)
			}

			m.migrateKey(ctx, key) // Basic processing

			if m.debug {
				log.Printf("[MIGRATION] Key processed: %s (took %v)",
					key, time.Since(startTime))
			}

		case <-ctx.Done():
			if m.debug {
				log.Printf("[MIGRATION] Worker stopped: context cancelled")
			}
			return
		}
	}
}

func (m *MigrationManager) processMigrations(ctx context.Context) {
	start := time.Now()
	defer func() {
		m.analytics.migrationTime.Observe(time.Since(start).Seconds())
	}()

	frequencyMap := m.analytics.GetFrequencyPerMinute()
	if frequencyMap == nil {
		return
	}

	for key, freq := range frequencyMap {
		currentLayer := m.getCurrentLayer(ctx, key)
		if currentLayer == 0 {
			continue // Skip keys in hot layer
		}
		targetLayer := m.selectTargetLayerIndex(freq)
		if targetLayer != -1 {
			select {
			case m.migrationQueue <- key:
				if m.debug {
					log.Printf("[MIGRATION] Key %s queued for migration", key)
				}
			default:
				m.analytics.migrationCount.WithLabelValues("queue_full").Inc()
				log.Printf("[MIGRATION] Migration queue full, dropping key %s", key)
			}
		}
	}
}

func (m *MigrationManager) migrateKey(ctx context.Context, key string) {
	if m.debug {
		log.Printf("[MIGRATION] Starting migration for key: %s", key)
	}
	start := time.Now()
	defer func() {
		m.analytics.migrationTime.Observe(time.Since(start).Seconds())
		if m.debug {
			log.Printf("[MIGRATION] Finished migration for key: %s", key)
		}
	}()

	// Determine the current key layer
	currentLayer := m.getCurrentLayer(ctx, key)
	freq := m.analytics.GetFrequency(key)
	targetLayer := m.selectTargetLayerIndex(freq)

	// If the key is already in the target or hotter layer, update the TTL and exit
	if currentLayer != -1 && currentLayer <= targetLayer {
		newTTL := m.ttlManager.calculateAdaptiveTTL(freq)
		m.ttlManager.AdjustTTL(key, newTTL)
		if m.debug {
			log.Printf("[MIGRATION] Key=%s is already in layer %d (>= target %d). TTL updated to %d",
				key, currentLayer, targetLayer, newTTL)
		}
		return
	}

	if targetLayer == -1 {
		if m.debug {
			log.Printf("[MIGRATION] Key=%s: no target layer found", key)
		}
		return
	}

	newTTL := m.ttlManager.calculateAdaptiveTTL(freq)
	value, err := m.findKeyValue(ctx, key)
	if err != nil {
		log.Printf("[MIGRATION] Error finding key %s: %v", key, err)
		return
	}

	if err = m.migrateToLayer(ctx, key, value, targetLayer, newTTL); err != nil {
		log.Printf("[MIGRATION] Error migrating key %s: %v", key, err)
	} else if m.debug {
		log.Printf("[MIGRATION] Key=%s migrated to layer %d", key, targetLayer)
	}
}

// getCurrentLayer Method for determining the current key layer
func (m *MigrationManager) getCurrentLayer(ctx context.Context, key string) int {
	resultCh := make(chan int, 1) // Channel to return layer index
	var wg sync.WaitGroup

	for i, layer := range m.layers {
		wg.Add(1)
		go func(index int, l LayerInfo) {
			defer wg.Done()
			if _, err := l.Layer.Get(ctx, key); err == nil {
				select {
				case resultCh <- index:
				default:
				}
			}
		}(i, layer)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for index := range resultCh {
		return index
	}
	return -1
}

func (m *MigrationManager) selectTargetLayerIndex(freq int) int {
	// We choose the hottest layer, not the first suitable one
	for i := 0; i < len(m.thresholds); i++ { // Iterate from the hot layer (index 0)
		if freq >= m.thresholds[i] {
			return i
		}
	}
	return -1
}

func (m *MigrationManager) findKeyValue(ctx context.Context, key string) (string, error) {
	for i := len(m.layers) - 1; i >= 0; i-- {
		value, err := m.layers[i].Layer.Get(ctx, key)
		if err == nil {
			return value, nil
		}
	}
	// If not found in cache, check database
	value, err := m.db.Get(ctx, key)
	if err == nil {
		return value, nil
	}
	return "", errors.New("key not found in any layer or database")
}

func (m *MigrationManager) migrateToLayer(ctx context.Context, key, value string, targetLayerIndex int, ttl int64) error {
	if targetLayerIndex >= len(m.layers) {
		return errors.New("[MIGRATION] Invalid target layer")
	}
	if m.debug {
		log.Printf("[MIGRATION] Migrating key=%s to layer=%d (TTL=%d)",
			key, targetLayerIndex, ttl)
	}

	if err := m.layers[targetLayerIndex].Layer.Set(ctx, key, value, time.Duration(ttl)*time.Second); err != nil {
		log.Printf("[MIGRATION] Failed to set key=%s in layer=%d: %v", key, targetLayerIndex, err)
		return err
	}
	m.ttlManager.AdjustTTL(key, ttl)
	if m.debug {
		log.Printf("[MIGRATION] Successfully migrated key=%s to layer=%d",
			key, targetLayerIndex)
	}
	return nil
}

func (m *MigrationManager) calculateInterval() time.Duration {
	interval := time.Duration(0)
	queueLen := len(m.migrationQueue)

	switch {
	case queueLen > 1000:
		interval = 500 * time.Millisecond
	case queueLen > 500:
		interval = 1 * time.Second
	default:
		interval = 5 * time.Second
	}

	if m.debug {
		log.Printf("[MIGRATION] New interval: %v (queue size: %d)", interval, queueLen)
	}
	return interval
}
