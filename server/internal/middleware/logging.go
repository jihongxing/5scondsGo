// Package middleware provides HTTP middleware components.
package middleware

import (
	"time"

	"github.com/fiveseconds/server/pkg/logger"
	"github.com/fiveseconds/server/pkg/metrics"
	"github.com/fiveseconds/server/pkg/trace"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// TraceIDHeader is the HTTP header for trace ID propagation
	TraceIDHeader = "X-Trace-ID"
)

// RequestLogging creates a middleware that logs HTTP requests with trace IDs
func RequestLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get or generate trace ID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = trace.NewTraceID()
		}

		// Inject trace ID into context
		ctx := trace.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// Set response header for trace ID propagation
		c.Header(TraceIDHeader, traceID)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record metrics
		metrics.RecordHTTPRequest(c.Request.Method, path, status, latency)

		// Build log fields
		fields := []zap.Field{
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add query string if present
		if c.Request.URL.RawQuery != "" {
			fields = append(fields, zap.String("query", c.Request.URL.RawQuery))
		}

		// Log based on status code
		if status >= 500 {
			log.WithContext(ctx).Error("request completed with server error", fields...)
		} else if status >= 400 {
			log.WithContext(ctx).Warn("request completed with client error", fields...)
		} else {
			log.WithContext(ctx).Info("request completed", fields...)
		}
	}
}

// RecoveryWithLogging creates a recovery middleware that logs panics with trace IDs
func RecoveryWithLogging(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				traceID := trace.GetTraceID(ctx)

				log.WithContext(ctx).Error("panic recovered",
					zap.String("trace_id", traceID),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Stack("stacktrace"),
				)

				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
