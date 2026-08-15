package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

		slog.Info("Incoming HTTP Request",
			slog.String("trace_id", traceID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("client_ip", c.ClientIP()),
		)

		c.Next()

		latency := time.Since(startTime).Milliseconds()
		slog.Info("HTTP Request Completed",
			slog.String("trace_id", traceID),
			slog.Int64("latency_ms", latency),
			slog.Int("status_code", c.Writer.Status()),
		)
	}
}
