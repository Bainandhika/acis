package cache

import (
	"testing"
	"time"
)

func TestTokenBucketLimiter(t *testing.T) {
	limiter := NewTokenBucketLimiter(1.0, 2.0, 1*time.Minute)
	defer limiter.Close()

	key := "127.0.0.1"

	if !limiter.Allow(key) {
		t.Fatal("expected first request to be allowed")
	}
	if !limiter.Allow(key) {
		t.Fatal("expected second request to be allowed (burst size 2)")
	}
	if limiter.Allow(key) {
		t.Fatal("expected third request to be rate limited")
	}
}
