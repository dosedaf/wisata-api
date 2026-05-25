package utils

import (
	"sync"
	"time"
)

type cacheItem struct {
	value      []byte
	expiration int64
}

// MemoryCache adalah struktur in-memory key-value store dasar
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]cacheItem),
	}
}

func (m *MemoryCache) Set(key string, value []byte, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(duration).UnixNano(),
	}
}

func (m *MemoryCache) Get(key string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, found := m.items[key]
	if !found {
		return nil, false
	}

	if time.Now().UnixNano() > item.expiration {
		return nil, false
	}

	return item.value, true
}
