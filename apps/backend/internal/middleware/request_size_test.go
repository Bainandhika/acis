package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestSizeLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestSizeLimiter(10)) // Max 10 bytes

	r.POST("/upload", func(c *gin.Context) {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(c.Request.Body)
		if err != nil {
			c.String(http.StatusRequestEntityTooLarge, "payload too large")
			return
		}
		c.String(http.StatusOK, "ok")
	})

	// Payload <= 10 bytes -> OK
	req1, _ := http.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("12345"))
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("Small payload expected 200, got %d", w1.Code)
	}

	// Payload > 10 bytes -> Exceeds limit
	req2, _ := http.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString("123456789012345"))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Large payload expected 413, got %d", w2.Code)
	}
}
