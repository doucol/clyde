// A simple in-memory cache that supports expiry
// https://www.alexedwards.net/blog/implementing-an-in-memory-cache-in-go
package cache

import (
	"math"
	"sync"
	"time"
)

const forever = time.Duration(math.MaxInt64)

// item represents a cache item with a value and an expiration time.
type item[V any] struct {
	value  V
	expiry time.Time
}

// isExpired checks if the cache item has expired.
func (i item[V]) isExpired() bool {
	return time.Now().After(i.expiry)
}

// Cache is a generic cache implementation with support for time-to-live
// (TTL) expiration.
type Cache[K comparable, V any] struct {
	items map[K]item[V] // The map storing cache items.
	mu    sync.RWMutex  // Mutex for controlling concurrent access to the cache.
}

// New creates a new Cache instance and starts a goroutine to periodically
// remove expired items every 5 seconds.
func New[K comparable, V any]() *Cache[K, V] {
	c := &Cache[K, V]{items: make(map[K]item[V])}

	go func() {
		for range time.Tick(5 * time.Second) {
			c.cull()
		}
	}()

	return c
}

// Iterate over the cache items and delete expired ones.
func (c *Cache[K, V]) cull() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, item := range c.items {
		if item.isExpired() {
			delete(c.items, key)
		}
	}
}

// Set adds a new item to the cache with the specified key, value
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetTTL(key, value, forever)
}

// Set adds a new item to the cache with the specified key, value, and
// time-to-live (TTL).
func (c *Cache[K, V]) SetTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item[V]{
		value:  value,
		expiry: time.Now().Add(ttl),
	}
}

// Get retrieves the value associated with the given key from the cache.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		// If the key is not found, return the zero value for V and false.
		return item.value, false
	}

	if item.isExpired() {
		// If the item has expired, remove it from the cache and return the
		// value and false.
		go c.Remove(key)
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}

// Remove removes the item with the specified key from the cache.
func (c *Cache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Delete the item with the given key from the cache.
	delete(c.items, key)
}

// Pop removes and returns the item with the specified key from the cache.
func (c *Cache[K, V]) Pop(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, found := c.items[key]
	if !found {
		// If the key is not found, return the zero value for V and false.
		return item.value, false
	}

	// If the key is found, delete the item from the cache.
	delete(c.items, key)

	if item.isExpired() {
		// If the item has expired, return the value and false.
		return item.value, false
	}

	// Otherwise return the value and true.
	return item.value, true
}
