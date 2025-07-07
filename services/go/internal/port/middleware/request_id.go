package middleware

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

// RequestID middleware adds a unique request ID to each request context
// This is important for tracing requests through logs and various services
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if we already have a request ID (from upstream, like a gateway)
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// If no ID exists, generate a new UUID
			requestID = uuid.New().String()
		}

		// Add it to the response headers for tracking
		w.Header().Set("X-Request-ID", requestID)
		
		// Store it in context
		ctx := middleware.WithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
