package cache_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/Bainandhika/acis/apps/backend/internal/shared/cache"
)

func TestTokenBucketLimiter_Allow(t *testing.T) {
	limiter := cache.NewTokenBucketLimiter(2.0, 3.0, 1*time.Minute)
	defer limiter.Close()

	key := "192.168.1.1"

	// 3 burst allowed
	assert.True(t, limiter.Allow(key))
	assert.True(t, limiter.Allow(key))
	assert.True(t, limiter.Allow(key))

	// 4th request rejected immediately
	assert.False(t, limiter.Allow(key))

	// Wait 550ms -> 1 token refilled (2 rps * 0.5s = 1 token)
	time.Sleep(550 * time.Millisecond)
	assert.True(t, limiter.Allow(key))
}
