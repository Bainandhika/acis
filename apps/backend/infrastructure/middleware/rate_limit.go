package middleware

import (
	"net/http"

	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/gin-gonic/gin"
)

// NativeRateLimitMiddleware limits requests per IP using in-memory token bucket
func NativeRateLimitMiddleware(limiter *cache.TokenBucketLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.Header("Retry-After", "60")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded, please slow down",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
