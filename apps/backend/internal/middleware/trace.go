package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const TraceIDKey = "X-Transaction-ID"

// TraceID generates a unique transaction ID for every request
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if trace ID already exists in header (e.g., from API Gateway)
		traceID := c.GetHeader(TraceIDKey)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set trace ID in response header
		c.Header(TraceIDKey, traceID)
		
		// Store trace ID in Gin context
		c.Set(TraceIDKey, traceID)

		// Log the incoming request with trace ID
		log.Info().
			Str("trace_id", traceID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Msg("Incoming HTTP Request")

		c.Next()
	}
}