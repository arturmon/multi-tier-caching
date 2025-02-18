package storage

import (
	"sync"
	"time"
)

type MemoryStorage struct {
	data sync.Map
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (m *MemoryStorage) Get(key string) (interface{}, bool) {
	return m.data.Load(key)
}

func (m *MemoryStorage) Set(key string, value interface{}, ttl time.Duration) {
	m.data.Store(key, value)
	go func() {
		time.Sleep(ttl)
		m.data.Delete(key)
	}()
}

func (m *MemoryStorage) Delete(key string) {
	m.data.Delete(key)
}
