package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/pkg/logger"
)

// HealthStatus contains the health status information
type HealthStatus struct {
	Status      string            `json:"status"`       // Overall status: "up", "down", or "degraded"
	Version     string            `json:"version"`      // Application version
	Environment string            `json:"environment"`  // Deployment environment
	Timestamp   time.Time         `json:"timestamp"`    // Current server time
	Uptime      string            `json:"uptime"`       // Service uptime
	Components  map[string]Status `json:"components"`   // Status of individual components
}

// Status represents the status of a component
type Status struct {
	Status  string `json:"status"`  // Component status: "up", "down", or "degraded"
	Message string `json:"message"` // Optional status message
}

// HealthHandler handles health check requests
type HealthHandler struct {
	log          logger.Logger
	dbPool       *pgxpool.Pool
	startTime    time.Time
	version      string
	environment  string
	dependencies map[string]func() (bool, string)
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(
	log logger.Logger, 
	dbPool *pgxpool.Pool, 
	version string, 
	environment string,
) *HealthHandler {
	handler := &HealthHandler{
		log:         log,
		dbPool:      dbPool,
		startTime:   time.Now(),
		version:     version,
		environment: environment,
		dependencies: make(map[string]func() (bool, string)),
	}
	
	// Register database health check
	handler.RegisterDependency("database", handler.checkDatabase)
	
	return handler
}

// RegisterDependency registers a new dependency health check
func (h *HealthHandler) RegisterDependency(name string, checkFunc func() (bool, string)) {
	h.dependencies[name] = checkFunc
}

// HandleHealth handles health check requests
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// Basic health check returns 200 OK without details
	if r.URL.Query().Get("detail") == "" {
		response.SendWithStatus(w, r, map[string]string{"status": "up"}, http.StatusOK)
		return
	}
	
	// Detailed health check
	status := h.buildHealthStatus()
	response.SendWithStatus(w, r, status, h.determineStatusCode(status))
}

// HandleReadiness handles readiness check requests
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	status := h.buildHealthStatus()
	
	// Set the status code based on overall health
	statusCode := h.determineStatusCode(status)
	response.SendWithStatus(w, r, status, statusCode)
}

// HandleLiveness handles liveness check requests
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	// Liveness only checks if the service is running, not dependencies
	response.SendWithStatus(w, r, map[string]string{
		"status": "up",
		"uptime": formatUptime(time.Since(h.startTime)),
	}, http.StatusOK)
}

// buildHealthStatus builds the complete health status
func (h *HealthHandler) buildHealthStatus() HealthStatus {
	components := make(map[string]Status)
	overallStatus := "up"
	
	// Check all registered dependencies
	for name, checkFunc := range h.dependencies {
		isUp, message := checkFunc()
		
		compStatus := "up"
		if !isUp {
			compStatus = "down"
			if overallStatus == "up" {
				overallStatus = "degraded"
			}
		}
		
		components[name] = Status{
			Status:  compStatus,
			Message: message,
		}
	}
	
	return HealthStatus{
		Status:      overallStatus,
		Version:     h.version,
		Environment: h.environment,
		Timestamp:   time.Now(),
		Uptime:      formatUptime(time.Since(h.startTime)),
		Components:  components,
	}
}

// determineStatusCode determines the HTTP status code based on health status
func (h *HealthHandler) determineStatusCode(status HealthStatus) int {
	switch status.Status {
	case "up":
		return http.StatusOK
	case "degraded":
		return http.StatusOK // Still functioning but with degraded performance
	default:
		return http.StatusServiceUnavailable
	}
}

// checkDatabase checks the database connection
func (h *HealthHandler) checkDatabase() (bool, string) {
	// Use a short timeout context for health checks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Simple ping to check connection
	err := h.dbPool.Ping(ctx)
	if err != nil {
		h.log.Error("Database health check failed", err)
		return false, "Database connection failed"
	}
	
	return true, "Connected"
}

// formatUptime formats the uptime duration as a human-readable string
func formatUptime(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	
	return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
}
