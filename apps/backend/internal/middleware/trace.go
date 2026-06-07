package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const TraceIDKey = "X-Transaction-ID"

// TraceID generates a unique transaction ID for every request
// and injects it into the standard context.Context so DB wrappers can use it.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 1. Check if trace ID already exists in header (e.g., from API Gateway or Frontend)
		traceID := c.GetHeader(TraceIDKey)
		if traceID == "" {
			// If not, generate a new UUID
			traceID = uuid.New().String()
		}

		// 2. Set trace ID in response header (so frontend/bot can log it too)
		c.Header(TraceIDKey, traceID)

		// 3. CRITICAL: Inject trace ID into the standard context.Context
		// This is how our DB wrapper (Commit 3) will read the trace ID later!
		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		// 4. Log the incoming request
		log.Info().
			Str("trace_id", traceID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Msg("Incoming HTTP Request")

		// Process request
		c.Next()

		// 5. Log the response latency
		latency := time.Since(startTime).Milliseconds()
		log.Info().
			Str("trace_id", traceID).
			Int64("latency_ms", latency).
			Int("status_code", c.Writer.Status()).
			Msg("HTTP Request Completed")
	}
}
