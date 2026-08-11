package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware injects a request context deadline to prevent long-hanging requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
