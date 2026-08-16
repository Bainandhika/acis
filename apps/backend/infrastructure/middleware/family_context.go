package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/gin-gonic/gin"
)

// FamilyContextMiddleware queries the user's family_id and role from the DB
// and injects them into the gin context. Must run AFTER AuthMiddleware.
func FamilyContextMiddleware(db *database.AppDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		uidStr, ok := userID.(string)
		if !ok || uidStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user identity"})
			c.Abort()
			return
		}

		var familyID string
		var role string
		query := `SELECT family_id, role FROM family_members WHERE user_id = $1 LIMIT 1`
		// Use c.Request.Context() to maintain distributed tracing context and request cancellation
		err := db.QueryRowContext(c.Request.Context(), query, uidStr).Scan(&familyID, &role)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusForbidden, gin.H{"error": "user must join a family first"})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve family context"})
			c.Abort()
			return
		}

		c.Set("family_id", familyID)
		c.Set("user_role", role) // Override the JWT role with DB-sourced role (more accurate)
		c.Next()
	}
}
