package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		contextRole    string
		setRole        bool
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Admin allowed on admin route",
			contextRole:    "admin",
			setRole:        true,
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Member forbidden on admin route",
			contextRole:    "member",
			setRole:        true,
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Role missing in context",
			setRole:        false,
			allowedRoles:   []string{"admin"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, r := gin.CreateTestContext(w)

			r.GET("/protected", func(c *gin.Context) {
				if tt.setRole {
					c.Set("user_role", tt.contextRole)
				}
				c.Next()
			}, RequireRole(tt.allowedRoles...), func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})

			c.Request, _ = http.NewRequest(http.MethodGet, "/protected", nil)
			r.ServeHTTP(w, c.Request)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
