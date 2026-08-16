package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.GET("/secure", func(c *gin.Context) {
		c.String(http.StatusOK, "secure content")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/secure", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	headers := map[string]string{
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=(), payment=()",
	}

	for k, expectedVal := range headers {
		if val := w.Header().Get(k); val != expectedVal {
			t.Errorf("header %s: expected '%s', got '%s'", k, expectedVal, val)
		}
	}
}
