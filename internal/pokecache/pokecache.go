package pokecache

import (
	"sync"
	"time"
)

// cacheEntry holds a single cached value along with the time it was created.
type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache is a thread-safe in-memory cache with automatic expiration.
type Cache struct {
	entries map[string]cacheEntry
	mu      sync.Mutex
}

// NewCache creates a new Cache and starts a background goroutine that
// periodically removes entries older than the given interval.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

// Add stores val under key, recording the current time as its creation time.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

// Get retrieves the value stored under key. The second return value is
// false if no entry was found for that key.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

// reapLoop runs forever, removing entries older than interval each time
// the ticker fires.
func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if time.Since(entry.createdAt) > interval {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}