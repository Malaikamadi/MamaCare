package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/pkg/logger"
)

// Middleware represents a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain represents a series of middleware that can be applied to an http.Handler
type Chain struct {
	middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) Chain {
	return Chain{
		middlewares: append([]Middleware(nil), middlewares...),
	}
}

// Then applies the middleware chain to an http.Handler
func (c Chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}

	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}

	return h
}

// Append creates a new Chain by appending the given middleware
// to the existing ones
func (c Chain) Append(middlewares ...Middleware) Chain {
	newMiddlewares := make([]Middleware, len(c.middlewares)+len(middlewares))
	copy(newMiddlewares, c.middlewares)
	copy(newMiddlewares[len(c.middlewares):], middlewares)

	return Chain{middlewares: newMiddlewares}
}

// ResponseWriter is a wrapper around http.ResponseWriter that provides
// status code tracking for logging and metrics
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

// WriteHeader captures the status code and calls the underlying WriteHeader
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the number of bytes written
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Status returns the HTTP status code
func (rw *ResponseWriter) Status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

// BytesWritten returns the number of bytes written
func (rw *ResponseWriter) BytesWritten() int64 {
	return rw.written
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		ctx := SetRequestID(r.Context(), requestID)
		w.Header().Set("X-Request-ID", requestID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs information about each HTTP request
func LoggingMiddleware(log logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture the status code
			rw := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Process the request
			next.ServeHTTP(rw, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Get request ID from context (if available)
			requestID := GetRequestID(r.Context())
			
			// Log request details with different levels based on status code
			if rw.Status() >= 500 {
				log.Error("Server error",
					logger.Field{Key: "method", Value: r.Method},
					logger.Field{Key: "path", Value: r.URL.Path},
					logger.Field{Key: "status", Value: rw.Status()},
					logger.Field{Key: "duration_ms", Value: duration.Milliseconds()},
					logger.Field{Key: "request_id", Value: requestID},
					logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
				)
			} else if rw.Status() >= 400 {
				log.Warn("Client error",
					logger.Field{Key: "method", Value: r.Method},
					logger.Field{Key: "path", Value: r.URL.Path},
					logger.Field{Key: "status", Value: rw.Status()},
					logger.Field{Key: "duration_ms", Value: duration.Milliseconds()},
					logger.Field{Key: "request_id", Value: requestID},
					logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
				)
			} else {
				log.Info("Request processed",
					logger.Field{Key: "method", Value: r.Method},
					logger.Field{Key: "path", Value: r.URL.Path},
					logger.Field{Key: "status", Value: rw.Status()},
					logger.Field{Key: "duration_ms", Value: duration.Milliseconds()},
					logger.Field{Key: "request_id", Value: requestID},
					logger.Field{Key: "bytes", Value: rw.BytesWritten()},
				)
			}
		})
	}
}

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware(log logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					requestID := GetRequestID(r.Context())
					
					log.Error("Panic recovered in HTTP handler",
						logger.Field{Key: "error", Value: rec},
						logger.Field{Key: "request_id", Value: requestID},
						logger.Field{Key: "path", Value: r.URL.Path},
						logger.Field{Key: "method", Value: r.Method},
					)
					
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds Cross-Origin Resource Sharing headers
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Check if the origin is allowed
			allowed := false
			if len(allowedOrigins) == 0 || allowedOrigins[0] == "*" {
				allowed = true
			} else {
				for _, allowedOrigin := range allowedOrigins {
					if origin == allowedOrigin {
						allowed = true
						break
					}
				}
			}
			
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "300")
			}
			
			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// CommonMiddleware returns a Chain with commonly used middleware
func CommonMiddleware(log logger.Logger) Chain {
	return NewChain(
		RequestIDMiddleware,
		LoggingMiddleware(log),
		RecoveryMiddleware(log),
		CORSMiddleware([]string{"*"}),
	)
}
