package repository

import (
	"strings"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
	DeleteByPrefix(prefix string)
	AcquireLock(key string, ttl time.Duration) bool
	ReleaseLock(key string)
}

type cacheEntry struct {
	Value     []byte
	ExpiresAt time.Time
}

type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[string]cacheEntry),
	}
}

func (c *InMemoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		c.Delete(key)
		return nil, false
	}
	value := make([]byte, len(entry.Value))
	copy(value, entry.Value)
	return value, true
}

func (c *InMemoryCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cloned := make([]byte, len(value))
	copy(cloned, value)

	entry := cacheEntry{Value: cloned}
	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}
	c.entries[key] = entry
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

func (c *InMemoryCache) DeleteByPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.entries {
		if strings.HasPrefix(key, prefix) {
			delete(c.entries, key)
		}
	}
}

func (c *InMemoryCache) AcquireLock(key string, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if ok && (entry.ExpiresAt.IsZero() || time.Now().Before(entry.ExpiresAt)) {
		return false
	}

	c.entries[key] = cacheEntry{
		Value:     []byte("1"),
		ExpiresAt: time.Now().Add(ttl),
	}
	return true
}

func (c *InMemoryCache) ReleaseLock(key string) {
	c.Delete(key)
}
