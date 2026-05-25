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

// NewMemoryCache membuat instance cache baru
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]cacheItem),
	}
}

// Set menyimpan data ke dalam memori dengan batas waktu tertentu
func (m *MemoryCache) Set(key string, value []byte, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items[key] = cacheItem{
		value:      value,
		expiration: time.Now().Add(duration).UnixNano(),
	}
}

// Get mengambil data dari memori. Mengembalikan false jika tidak ada atau sudah expired
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
