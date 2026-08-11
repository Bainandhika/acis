package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*clientLimiter
	rate    rate.Limit
	burst   int
	ttl     time.Duration
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		clients: make(map[string]*clientLimiter),
		rate:    r,
		burst:   b,
		ttl:     10 * time.Minute,
	}

	go i.cleanupStaleVisitors()
	return i
}

func (i *IPRateLimiter) getVisitor(key string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	v, exists := i.clients[key]
	if !exists {
		limiter := rate.NewLimiter(i.rate, i.burst)
		i.clients[key] = &clientLimiter{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (i *IPRateLimiter) cleanupStaleVisitors() {
	for {
		time.Sleep(3 * time.Minute)
		i.mu.Lock()
		for key, visitor := range i.clients {
			if time.Since(visitor.lastSeen) > i.ttl {
				delete(i.clients, key)
			}
		}
		i.mu.Unlock()
	}
}

func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			if uidStr, ok := userID.(string); ok && uidStr != "" {
				key = "user:" + uidStr
			}
		}

		lim := limiter.getVisitor(key)
		if !lim.Allow() {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please slow down",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func NativeRateLimitMiddleware(limiter *cache.TokenBucketLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			if uidStr, ok := userID.(string); ok && uidStr != "" {
				key = "user:" + uidStr
			}
		}

		if !limiter.Allow(key) {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please slow down",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
