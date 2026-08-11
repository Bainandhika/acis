package cache

import (
	"sync"
	"time"
)

type cacheItem[V any] struct {
	value     V
	expiresAt time.Time
}

type TTLCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]cacheItem[V]
	stop  chan struct{}
}

func NewTTLCache[K comparable, V any](cleanupInterval time.Duration) *TTLCache[K, V] {
	if cleanupInterval <= 0 {
		cleanupInterval = 1 * time.Minute
	}
	c := &TTLCache[K, V]{
		items: make(map[K]cacheItem[V]),
		stop:  make(chan struct{}),
	}
	go c.startPurgeTicker(cleanupInterval)
	return c
}

func (c *TTLCache[K, V]) Set(key K, val V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem[V]{
		value:     val,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[key]
	if !exists || time.Now().After(item.expiresAt) {
		var zero V
		return zero, false
	}
	return item.value, true
}

func (c *TTLCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *TTLCache[K, V]) startPurgeTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for k, item := range c.items {
				if now.After(item.expiresAt) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

func (c *TTLCache[K, V]) Close() {
	select {
	case <-c.stop:
		// already closed
	default:
		close(c.stop)
	}
}
