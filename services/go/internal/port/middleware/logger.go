package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/mamacare/services/pkg/logger"
)

// Logger creates a middleware that logs request information
func Logger(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Create a request-scoped logger with the request ID
			reqID := middleware.GetReqID(r.Context())
			requestLog := log.With().Str("request_id", reqID).Logger()
			
			// Store start time for duration calculation
			start := time.Now()
			
			// Record request details at debug level
			requestLog.Debug().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Msg("Request received")
			
			// Create a wrapper for the response writer to capture the status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			
			// Call the next handler
			next.ServeHTTP(ww, r)
			
			// Calculate request duration
			duration := time.Since(start)
			
			// Log the response details at info level
			level := requestLog.Info()
			if ww.Status() >= 400 {
				level = requestLog.Warn()
			}
			if ww.Status() >= 500 {
				level = requestLog.Error()
			}
			
			level.
				Int("status", ww.Status()).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("duration", duration).
				Int("bytes", ww.BytesWritten()).
				Msg("Request completed")
		}
		return http.HandlerFunc(fn)
	}
}
