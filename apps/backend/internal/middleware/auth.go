package middleware

import (
	"net/http"
	"strings"

	"github.com/Bainandhika/acis/apps/backend/pkg/auth"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT and injects user info into context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// 2. Check format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		// 3. Validate Token
		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 4. Inject User Info into Context (So handlers can use it)
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next() // Continue to the next handler
	}
}
