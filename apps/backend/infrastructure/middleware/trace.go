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
// and injects it into context.Context.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		traceID := c.GetHeader(TraceIDKey)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Header(TraceIDKey, traceID)

		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		log.Info().
			Str("trace_id", traceID).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("client_ip", c.ClientIP()).
			Msg("Incoming HTTP Request")

		c.Next()

		latency := time.Since(startTime).Milliseconds()
		log.Info().
			Str("trace_id", traceID).
			Int64("latency_ms", latency).
			Int("status_code", c.Writer.Status()).
			Msg("HTTP Request Completed")
	}
}
