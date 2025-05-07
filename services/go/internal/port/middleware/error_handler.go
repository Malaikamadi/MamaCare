package middleware

import (
	"fmt"
	"net/http"

	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// ErrorHandlerMiddleware wraps handlers with standard error handling
type ErrorHandlerMiddleware struct {
	log logger.Logger
}

// NewErrorHandlerMiddleware creates a new error handler middleware
func NewErrorHandlerMiddleware(log logger.Logger) *ErrorHandlerMiddleware {
	return &ErrorHandlerMiddleware{
		log: log,
	}
}

// Handler wraps an HTTP handler with error handling
func (m *ErrorHandlerMiddleware) Handler(handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get request ID for correlation
		requestID := GetRequestID(r.Context())
		
		// Execute the handler and catch any errors
		err := handler(w, r)
		if err != nil {
			// If it's a custom error, check the type
			if appErr, ok := err.(*errorx.Error); ok {
				// If it's a client error, log at warn level
				if errorx.HTTPStatusCode(appErr) < 500 {
					m.log.Warn("Client error in API request", 
						logger.Field{Key: "error", Value: appErr.Error()},
						logger.Field{Key: "error_type", Value: fmt.Sprintf("%d", appErr.GetType())})
				} else {
					m.log.Error("Server error in API request", err)
				}
			} else {
				m.log.Error("Unhandled error in API request", err)
			}
			
			// Get additional request details for logging
			method := r.Method
			path := r.URL.Path

			// Add request details to logs
			m.log.Debug("Request details for error",
				logger.Field{Key: "request_id", Value: requestID},
				logger.Field{Key: "method", Value: method},
				logger.Field{Key: "path", Value: path},
				logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
			)
			
			// Send error response
			statusCode := errorx.HTTPStatusCode(err)
			response := errorx.NewErrorResponse(err, requestID)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			w.Write([]byte(response.String()))
			return
		}
	}
}

// Wrap wraps a handler with error handling
func (m *ErrorHandlerMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
