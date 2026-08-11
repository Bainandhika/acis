package cache

import (
	"testing"
	"time"
)

func TestTTLCache(t *testing.T) {
	cache := NewTTLCache[string, string](50 * time.Millisecond)
	defer cache.Close()

	cache.Set("key1", "val1", 100*time.Millisecond)
	cache.Set("key2", "val2", 500*time.Millisecond)

	val, found := cache.Get("key1")
	if !found || val != "val1" {
		t.Fatalf("expected key1 to be val1, got %v (found: %v)", val, found)
	}

	time.Sleep(150 * time.Millisecond)

	_, found = cache.Get("key1")
	if found {
		t.Fatalf("expected key1 to be expired")
	}

	val, found = cache.Get("key2")
	if !found || val != "val2" {
		t.Fatalf("expected key2 to still be val2, got %v (found: %v)", val, found)
	}
}
