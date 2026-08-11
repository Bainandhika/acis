package cache

import (
	"math"
	"sync"
	"time"
)

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

func newTokenBucket(rps float64, burst float64) *tokenBucket {
	return &tokenBucket{
		tokens:     burst,
		maxTokens:  burst,
		refillRate: rps,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens = math.Min(b.maxTokens, b.tokens+(elapsed*b.refillRate))
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false
}

type TokenBucketLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*tokenBucket
	rps     float64
	burst   float64
	stop    chan struct{}
	ttl     time.Duration
}

func NewTokenBucketLimiter(rps float64, burst float64, cleanupInterval time.Duration) *TokenBucketLimiter {
	if rps <= 0 {
		rps = 10.0
	}
	if burst <= 0 {
		burst = 20.0
	}
	if cleanupInterval <= 0 {
		cleanupInterval = 5 * time.Minute
	}

	l := &TokenBucketLimiter{
		buckets: make(map[string]*tokenBucket),
		rps:     rps,
		burst:   burst,
		stop:    make(chan struct{}),
		ttl:     10 * time.Minute,
	}

	go l.startPurgeTicker(cleanupInterval)
	return l
}

func (l *TokenBucketLimiter) Allow(key string) bool {
	l.mu.RLock()
	bucket, exists := l.buckets[key]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		bucket, exists = l.buckets[key]
		if !exists {
			bucket = newTokenBucket(l.rps, l.burst)
			l.buckets[key] = bucket
		}
		l.mu.Unlock()
	}

	return bucket.allow()
}

func (l *TokenBucketLimiter) startPurgeTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for k, b := range l.buckets {
				b.mu.Lock()
				if now.Sub(b.lastRefill) > l.ttl {
					delete(l.buckets, k)
				}
				b.mu.Unlock()
			}
			l.mu.Unlock()
		case <-l.stop:
			return
		}
	}
}

func (l *TokenBucketLimiter) Close() {
	select {
	case <-l.stop:
		// already closed
	default:
		close(l.stop)
	}
}
