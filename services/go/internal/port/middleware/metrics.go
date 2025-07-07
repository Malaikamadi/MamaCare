package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mamacare/services/pkg/logger"
	"github.com/mamacare/services/pkg/metrics"
)

// Metrics middleware collects API usage metrics
func Metrics(metricsClient metrics.Client, log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Record start time for latency tracking
			startTime := time.Now()
			
			// Get simplified route pattern
			routePattern := getSimplifiedPath(r.URL.Path)
			
			// Create response writer wrapper to capture status code
			ww := &ResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			// Call the next handler
			next.ServeHTTP(ww, r)
			
			// Calculate duration
			duration := time.Since(startTime)
			statusCode := ww.Status()
			
			// Classify status (success, client error, server error)
			var outcome string
			switch {
			case statusCode >= 500:
				outcome = "server_error"
			case statusCode >= 400:
				outcome = "client_error"
			default:
				outcome = "success"
			}
			
			// Log metrics at debug level
			log.Debug("API metrics collected",
				logger.Field{Key: "method", Value: r.Method},
				logger.Field{Key: "route", Value: routePattern},
				logger.Field{Key: "status", Value: statusCode},
				logger.Field{Key: "outcome", Value: outcome},
				logger.Field{Key: "duration_ms", Value: duration})
			
			// Record metrics if client is available
			if metricsClient != nil {
				// Track request count
				metricsClient.Count(
					"api.requests.count",
					1,
					[]string{
						"method:" + r.Method,
						"route:" + routePattern,
						"status:" + fmt.Sprintf("%d", statusCode),
						"outcome:" + outcome,
					},
					1.0,
				)
				
				// Track request duration
				metricsClient.Histogram(
					"api.requests.duration",
					float64(duration.Milliseconds()),
					[]string{
						"method:" + r.Method,
						"route:" + routePattern,
						"outcome:" + outcome,
					},
					1.0,
				)
				
				// Track response size
				metricsClient.Histogram(
					"api.responses.size",
					float64(ww.BytesWritten()),
					[]string{
						"method:" + r.Method,
						"route:" + routePattern,
					},
					1.0,
				)
			}
		})
	}
}

// getSimplifiedPath creates a simplified route pattern from a path
// by replacing numeric IDs with :id placeholders
func getSimplifiedPath(path string) string {
	// Split the path into segments
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for i, segment := range segments {
		// Check if segment looks like an ID (UUID or numeric)
		if (len(segment) > 8 && strings.Contains(segment, "-")) || isNumeric(segment) {
			segments[i] = ":id"
		}
	}
	
	// Join the segments back together
	if len(segments) == 0 || (len(segments) == 1 && segments[0] == "") {
		return "/"
	}
	
	return "/" + strings.Join(segments, "/")
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
