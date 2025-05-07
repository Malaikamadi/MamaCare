package hasura

import (
	"context"
)

// contextKey is a custom type for context keys
type contextKey string

// Context keys
const (
	requestIDKey contextKey = "request_id"
)

// getRequestID extracts request ID from the context
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}
	return ""
}

// withRequestID adds request ID to context
func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}
