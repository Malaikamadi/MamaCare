package middleware

import (
	"net/http"
	"strings"

	"github.com/mamacare/services/pkg/logger"
)

// CORSConfig defines settings for CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of origins that can access the resource
	// Use ["*"] to allow any origin
	AllowedOrigins []string
	
	// AllowedMethods is a list of HTTP methods allowed
	AllowedMethods []string
	
	// AllowedHeaders is a list of HTTP headers allowed
	AllowedHeaders []string
	
	// ExposedHeaders are the headers exposed to the browser
	ExposedHeaders []string
	
	// AllowCredentials indicates if requests can include user credentials like cookies
	AllowCredentials bool
	
	// MaxAge indicates how long (in seconds) the results of a preflight request can be cached
	MaxAge int
}

// DefaultCORSConfig returns a default configuration for development
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID", "X-Requested-With"},
		ExposedHeaders:   []string{"Link", "X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(config CORSConfig, log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Skip if no Origin header is present (same origin request)
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}
			
			// Check if the origin is allowed
			allowed := false
			for _, allowedOrigin := range config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if !allowed {
				log.Debug("CORS request from disallowed origin", 
					logger.Field{Key: "origin", Value: origin})
				next.ServeHTTP(w, r)
				return
			}
			
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			
			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if config.MaxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", string(config.MaxAge))
				}
				if len(config.ExposedHeaders) > 0 {
					w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			
			// Continuing with the actual request
			next.ServeHTTP(w, r)
		})
	}
}
