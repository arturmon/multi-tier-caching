package cachelib

import (
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

type BloomFilter struct {
	filter *bloom.BloomFilter
	mu     sync.Mutex
}

func NewBloomFilter(size uint, hashFuncs uint) *BloomFilter {
	return &BloomFilter{
		filter: bloom.New(size, hashFuncs),
	}
}

func (b *BloomFilter) Add(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.filter.AddString(key)
}

func (b *BloomFilter) Exists(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.filter.TestString(key)
}
