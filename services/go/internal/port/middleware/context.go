package middleware

import (
	"context"
	"net/http"
)

// contextKey is a custom type for context keys
type contextKey string

// Context keys
const (
	requestIDKey contextKey = "request_id"
	userKey      contextKey = "user"
	roleKey      contextKey = "role"
	claimsKey    contextKey = "claims"
)

// RequestIDMiddleware adds a request ID to the context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get request ID from header or generate a new one
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// Add to response header
		w.Header().Set("X-Request-ID", requestID)
		
		// Store in context
		ctx := WithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID creates a new unique request ID
func generateRequestID() string {
	// Simple UUID v4 implementation would go here
	// For simplicity, we're using a placeholder
	return "req-" + GenerateRandomString(16)
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	// Simple random string implementation would go here
	// For simplicity, we're using a placeholder
	return "random-string"
}

// WithRequestID adds a request ID to a context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID gets the request ID from a context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithUser adds a user to a context
func WithUser(ctx context.Context, user interface{}) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// GetUser gets the user from a context
func GetUser(ctx context.Context) interface{} {
	return ctx.Value(userKey)
}

// WithRole adds a role to a context
func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKey, role)
}

// GetRole gets the role from a context
func GetRole(ctx context.Context) string {
	if role, ok := ctx.Value(roleKey).(string); ok {
		return role
	}
	return ""
}

// WithClaims adds claims to a context
func WithClaims(ctx context.Context, claims map[string]interface{}) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// GetClaims gets claims from a context
func GetClaims(ctx context.Context) map[string]interface{} {
	if claims, ok := ctx.Value(claimsKey).(map[string]interface{}); ok {
		return claims
	}
	return nil
}
