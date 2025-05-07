package http

import (
	"net/http"

	"github.com/mamacare/services/internal/port/handler"
	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/pkg/logger"
	"github.com/mamacare/services/pkg/metrics"
)

// CommonRouter implements shared router functionality
type CommonRouter struct {
	router        *http.ServeMux
	log           logger.Logger
	errorHandler  *middleware.ErrorHandlerMiddleware
	healthHandler *handler.HealthHandler
	metricsClient metrics.Client
}

// NewCommonRouter creates a new common router
func NewCommonRouter(
	log logger.Logger,
	healthHandler *handler.HealthHandler,
	metricsClient metrics.Client,
) *CommonRouter {
	r := http.NewServeMux()
	
	// Create the error handler middleware
	errorHandler := middleware.NewErrorHandlerMiddleware(log)
	
	return &CommonRouter{
		router:        r,
		log:           log,
		errorHandler:  errorHandler,
		healthHandler: healthHandler,
		metricsClient: metricsClient,
	}
}

// SetupMiddleware configures all common middleware
func (r *CommonRouter) SetupMiddleware() {
	// NOTE: http.ServeMux doesn't support middleware directly
	// Middleware will need to be applied at the HTTP server level
	// or using a middleware wrapper around the final handler
	r.log.Info("Middleware setup is managed externally for standard ServeMux")
}

// commonHeaders adds standard headers to all responses
func (r *CommonRouter) commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		// Continue to the next handler
		next.ServeHTTP(w, req)
	})
}

// SetupHealthRoutes sets up health check routes
func (r *CommonRouter) SetupHealthRoutes() {
	router := r.router
	
	// Simple health check
	router.HandleFunc("/health", r.healthHandler.HandleHealth)
	
	// Detailed health checks
	router.HandleFunc("/health/readiness", r.healthHandler.HandleReadiness)
	router.HandleFunc("/health/liveness", r.healthHandler.HandleLiveness)
}

// SetupAPIRoutes sets up API routes with version prefixes
func (r *CommonRouter) SetupAPIRoutes(mountHandlers func(apiHandler http.Handler)) {
	// With standard ServeMux, we can't have subrouters
	// API handlers need to be registered with full paths
	// This is a simplified implementation
	mountHandlers(r.router)
}

// Handler returns the router as an http.Handler
func (r *CommonRouter) Handler() http.Handler {
	return r.router
}

// WrapHandler wraps a handler function with error handling
func (r *CommonRouter) WrapHandler(handler func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return r.errorHandler.Handler(handler)
}
