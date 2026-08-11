package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/Bainandhika/acis/apps/backend/internal/shared/cache"
)

func TestTTLCache_SetAndGet(t *testing.T) {
	c := cache.NewTTLCache[string, string](100 * time.Millisecond)
	defer c.Close()

	c.Set("key1", "val1", 500*time.Millisecond)
	val, ok := c.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "val1", val)

	time.Sleep(600 * time.Millisecond)
	_, ok = c.Get("key1")
	assert.False(t, ok)
}

func TestTTLCache_Delete(t *testing.T) {
	c := cache.NewTTLCache[string, int](1 * time.Minute)
	defer c.Close()

	c.Set("k", 42, 1*time.Minute)
	c.Delete("k")
	_, ok := c.Get("k")
	assert.False(t, ok)
}

func TestTTLCache_ConcurrentAccess(t *testing.T) {
	c := cache.NewTTLCache[int, int](50 * time.Millisecond)
	defer c.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c.Set(id, id*2, 200*time.Millisecond)
			val, ok := c.Get(id)
			if ok {
				assert.Equal(t, id*2, val)
			}
		}(i)
	}
	wg.Wait()
}
