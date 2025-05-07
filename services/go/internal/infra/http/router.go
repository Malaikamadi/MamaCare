package http

import (
	"net/http"

	"github.com/mamacare/services/internal/app/auth"
	appHandler "github.com/mamacare/services/internal/port/handler"
	appMiddleware "github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/internal/port/middleware"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/pkg/logger"
	"github.com/mamacare/services/pkg/metrics"
)

// Router configures and returns the HTTP router
type Router struct {
	authService     *auth.Service
	authHandler     *appHandler.AuthHandler
	hasuraWebhook   *appHandler.HasuraAuthWebhook
	authMiddleware  *appMiddleware.AuthMiddleware
	logger          logger.Logger
	metricsClient   metrics.Client
}

// NewRouter creates a new router
func NewRouter(
	authService *auth.Service,
	authHandler *appHandler.AuthHandler,
	hasuraWebhook *appHandler.HasuraAuthWebhook,
	authMiddleware *appMiddleware.AuthMiddleware,
	logger logger.Logger,
	metricsClient metrics.Client,
) *Router {
	return &Router{
		authService:     authService,
		authHandler:     authHandler,
		hasuraWebhook:   hasuraWebhook,
		authMiddleware:  authMiddleware,
		logger:          logger,
		metricsClient:   metricsClient,
	}
}

// Setup configures and returns the router with middleware
func (r *Router) Setup() http.Handler {
	// Create a new HTTP mux
	router := http.NewServeMux()
	
	// Create middleware chains for different route types
	// Common middleware for all routes
	commonChain := middleware.NewChain(
		middleware.RequestIDMiddleware,
		middleware.LoggingMiddleware(r.logger),
		middleware.RecoveryMiddleware(r.logger),
		middleware.CORSMiddleware([]string{"*"}), // Allow all origins for development
		middleware.Metrics(r.metricsClient, r.logger), // Add metrics collection
	)
	
	// Auth middleware chain for protected routes
	authChain := commonChain.Append(r.authMiddleware.Authenticate)
	
	// Health check route
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Simple OK response for health checks
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`)) 
	})
	
	// Public routes - no authentication
	router.HandleFunc("/auth/login", r.authHandler.Login)
	router.HandleFunc("/auth/refresh", r.authHandler.RefreshToken)
	router.HandleFunc("/hasura-webhook", r.hasuraWebhook.Authenticate)
	
	// User routes - require authentication
	router.Handle("/user/hasura-jwt", authChain.Then(
		http.HandlerFunc(r.authHandler.GenerateHasuraJWT),
	))
	
	// Protected routes for CHWs - require authentication and role check
	chwChain := authChain.Append(func(next http.Handler) http.Handler {
		return r.authMiddleware.RequireRole([]string{"CHW", "ADMIN"})(next)
	})
	
	// Add example CHW route - to be expanded
	router.Handle("/chw/profile", chwChain.Then(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for CHW profile handler
			response.Send(w, r, map[string]string{"message": "CHW profile endpoint"})
		}),
	))
	
	// Protected routes for mothers - require authentication and role check
	motherChain := authChain.Append(func(next http.Handler) http.Handler {
		return r.authMiddleware.RequireRole([]string{"MOTHER", "CHW", "CLINICIAN", "ADMIN"})(next)
	})
	
	// Add example mother route - to be expanded
	router.Handle("/mother/profile", motherChain.Then(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Placeholder for mother profile handler
			response.Send(w, r, map[string]string{"message": "Mother profile endpoint"})
		}),
	))
	
	// Log completion
	r.logger.Info("Router setup complete with middleware chains")
	
	// Apply common middleware to the entire router and return
	return commonChain.Then(router)
}