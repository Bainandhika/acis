package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func BotSecretMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if secret == "" {
			c.Next()
			return
		}
		incoming := c.GetHeader("X-Bot-Secret")
		if incoming != secret {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized bot request"})
			return
		}
		c.Next()
	}
}
