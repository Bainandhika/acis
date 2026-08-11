package middleware

import (
	"net/http"
	"strings"

	"github.com/Bainandhika/acis/apps/backend/shared/security"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT and injects user info into context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. Check Authorization Header "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 2. Fallback to HttpOnly Cookie "auth_token"
		if tokenString == "" {
			if cookieToken, err := c.Cookie("auth_token"); err == nil && cookieToken != "" {
				tokenString = cookieToken
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		// 3. Validate Token using provided jwtSecret
		claims, err := security.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 4. Inject User Info into Context
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}
