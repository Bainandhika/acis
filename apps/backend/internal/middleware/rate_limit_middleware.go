package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Bainandhika/acis/apps/backend/internal/shared/cache"
)

// NativeRateLimitMiddleware creates a Gin middleware that uses cache.TokenBucketLimiter
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
