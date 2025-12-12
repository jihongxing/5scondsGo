// Package trace provides distributed tracing context management.
package trace

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	// TraceIDKey is the context key for trace ID
	TraceIDKey contextKey = "trace_id"
	// SessionIDKey is the context key for WebSocket session ID
	SessionIDKey contextKey = "session_id"
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
)

// TraceContext holds tracing information for a request
type TraceContext struct {
	TraceID   string
	SessionID string
	UserID    int64
}

// NewTraceID generates a new unique trace ID using UUID v4
func NewTraceID() string {
	return uuid.New().String()
}

// NewSessionID generates a new unique session ID for WebSocket connections
func NewSessionID() string {
	return uuid.New().String()
}

// WithTraceID injects a trace ID into the context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID extracts the trace ID from context
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(TraceIDKey); v != nil {
		return v.(string)
	}
	return ""
}


// WithSessionID injects a session ID into the context
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// GetSessionID extracts the session ID from context
func GetSessionID(ctx context.Context) string {
	if v := ctx.Value(SessionIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// WithUserID injects a user ID into the context
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID extracts the user ID from context
func GetUserID(ctx context.Context) int64 {
	if v := ctx.Value(UserIDKey); v != nil {
		return v.(int64)
	}
	return 0
}

// WithTraceContext injects a complete trace context
func WithTraceContext(ctx context.Context, tc *TraceContext) context.Context {
	ctx = context.WithValue(ctx, TraceIDKey, tc.TraceID)
	if tc.SessionID != "" {
		ctx = context.WithValue(ctx, SessionIDKey, tc.SessionID)
	}
	if tc.UserID > 0 {
		ctx = context.WithValue(ctx, UserIDKey, tc.UserID)
	}
	return ctx
}

// GetTraceContext extracts the complete trace context from context
func GetTraceContext(ctx context.Context) *TraceContext {
	tc := &TraceContext{}
	tc.TraceID = GetTraceID(ctx)
	tc.SessionID = GetSessionID(ctx)
	tc.UserID = GetUserID(ctx)
	return tc
}
