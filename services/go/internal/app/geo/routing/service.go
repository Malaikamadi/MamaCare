package routing

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/mamacare/services/internal/app/geo/location"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// RoutePoint represents a point in a route
type RoutePoint struct {
	Location      model.Location `json:"location"`
	Name          string         `json:"name,omitempty"`
	Address       string         `json:"address,omitempty"`
	DistanceToNext float64       `json:"distance_to_next,omitempty"` // km
	Duration      int            `json:"duration,omitempty"`         // seconds
	ArrivalTime   string         `json:"arrival_time,omitempty"`
}

// Route represents a calculated route
type Route struct {
	Points         []RoutePoint   `json:"points"`
	TotalDistance  float64        `json:"total_distance"` // km
	TotalDuration  int            `json:"total_duration"` // seconds
	StartTime      string         `json:"start_time"`
	EndTime        string         `json:"end_time"`
}

// TransportMode represents a mode of transportation
type TransportMode string

const (
	// TransportModeDriving represents driving
	TransportModeDriving TransportMode = "driving"
	// TransportModeWalking represents walking
	TransportModeWalking TransportMode = "walking"
	// TransportModeBicycling represents bicycling
	TransportModeBicycling TransportMode = "bicycling"
)

// RouteOptions represents options for route calculation
type RouteOptions struct {
	TransportMode  TransportMode  `json:"transport_mode"`
	AvoidHighways  bool           `json:"avoid_highways,omitempty"`
	DepartureTime  time.Time      `json:"departure_time,omitempty"`
	Optimize       bool           `json:"optimize,omitempty"`
}

// Service provides routing functionality
type Service struct {
	log             logger.Logger
	locationService *location.Service
}

// NewService creates a new routing service
func NewService(
	log logger.Logger,
	locationService *location.Service,
) *Service {
	return &Service{
		log:             log,
		locationService: locationService,
	}
}

// CalculateRoute calculates a route between multiple points
func (s *Service) CalculateRoute(
	ctx context.Context,
	points []model.Location,
	options *RouteOptions,
) (*Route, error) {
	if len(points) < 2 {
		return nil, errorx.New(errorx.BadRequest, "Route requires at least 2 points")
	}

	// In a production environment, this would call a routing API like OSRM, Mapbox, etc.
	// For our MVP, we'll implement a simplified version using direct distance calculations

	// Default options if not provided
	if options == nil {
		options = &RouteOptions{
			TransportMode: TransportModeDriving,
			Optimize:      false,
		}
	}

	// Use current time if not specified
	departureTime := options.DepartureTime
	if departureTime.IsZero() {
		departureTime = time.Now()
	}

	// If optimize is true, reorder waypoints for shortest path (except start and end)
	orderedPoints := points
	if options.Optimize && len(points) > 3 {
		orderedPoints = s.optimizeWaypoints(points)
	}

	// Calculate route
	routePoints := make([]RoutePoint, len(orderedPoints))
	currentTime := departureTime
	totalDistance := 0.0
	totalDuration := 0

	for i, point := range orderedPoints {
		routePoint := RoutePoint{
			Location: point,
		}

		// If not the last point, calculate distance to next point
		if i < len(orderedPoints)-1 {
			nextPoint := orderedPoints[i+1]
			distance := s.locationService.CalculateDistance(
				point.Latitude, point.Longitude,
				nextPoint.Latitude, nextPoint.Longitude,
			)
			
			// Calculate duration based on transport mode
			duration := s.calculateDuration(distance, options.TransportMode)
			
			routePoint.DistanceToNext = distance
			routePoint.Duration = duration
			
			// Format arrival time
			routePoint.ArrivalTime = currentTime.Format(time.RFC3339)
			
			// Update current time for next point
			currentTime = currentTime.Add(time.Duration(duration) * time.Second)
			
			// Update totals
			totalDistance += distance
			totalDuration += duration
		} else {
			// Last point
			routePoint.ArrivalTime = currentTime.Format(time.RFC3339)
		}

		routePoints[i] = routePoint
	}

	// Create route
	route := &Route{
		Points:        routePoints,
		TotalDistance: totalDistance,
		TotalDuration: totalDuration,
		StartTime:     departureTime.Format(time.RFC3339),
		EndTime:       currentTime.Format(time.RFC3339),
	}

	return route, nil
}

// FindNearestFacility finds the nearest healthcare facility to a location
// This is a stub implementation for MVP - in a production environment,
// this would leverage a more sophisticated routing system
func (s *Service) FindNearestFacility(
	ctx context.Context,
	lat, lng float64,
	facilityType model.FacilityType,
) (*model.HealthcareFacility, *Route, error) {
	// This function would typically:
	// 1. Query for facilities of the given type
	// 2. Calculate routes to each facility
	// 3. Return the nearest one (by travel time, not just distance)
	
	// For the MVP, we're returning a stub indicating this needs to be
	// implemented with the facility repository
	return nil, nil, errorx.New(errorx.Internal, "Pending implementation with facility repository")
}

// CalculateVisitRoute calculates an optimized route for CHW visit planning
func (s *Service) CalculateVisitRoute(
	ctx context.Context,
	chwLocation model.Location,
	motherLocations []model.Location,
	options *RouteOptions,
) (*Route, error) {
	// Create a list of all points including CHW start location
	allPoints := make([]model.Location, 0, len(motherLocations)+2)
	
	// Start at CHW location
	allPoints = append(allPoints, chwLocation)
	
	// Add all mother locations
	allPoints = append(allPoints, motherLocations...)
	
	// End at CHW location (return to base)
	allPoints = append(allPoints, chwLocation)
	
	// Always optimize for visit planning
	if options == nil {
		options = &RouteOptions{
			TransportMode: TransportModeDriving,
			Optimize:      true,
		}
	} else {
		options.Optimize = true
	}
	
	// Calculate optimized route
	return s.CalculateRoute(ctx, allPoints, options)
}

// optimizeWaypoints reorders waypoints for shortest path (except first and last)
func (s *Service) optimizeWaypoints(points []model.Location) []model.Location {
	// Keep first and last points fixed
	start := points[0]
	end := points[len(points)-1]
	
	// Points to optimize
	waypoints := points[1 : len(points)-1]
	
	// If only one or zero waypoints, no optimization needed
	if len(waypoints) <= 1 {
		return points
	}
	
	// For MVP, use a simple greedy algorithm (nearest neighbor)
	optimized := make([]model.Location, 0, len(points))
	optimized = append(optimized, start)
	
	// Create a copy of waypoints to work with
	remaining := make([]model.Location, len(waypoints))
	copy(remaining, waypoints)
	
	// Current position
	current := start
	
	// Keep finding the nearest point until all waypoints are visited
	for len(remaining) > 0 {
		nextIdx := s.findNearestPointIndex(current, remaining)
		next := remaining[nextIdx]
		
		// Add to optimized route
		optimized = append(optimized, next)
		
		// Remove from remaining
		remaining = append(remaining[:nextIdx], remaining[nextIdx+1:]...)
		
		// Update current position
		current = next
	}
	
	// Add the end point
	optimized = append(optimized, end)
	
	return optimized
}

// findNearestPointIndex finds the index of the nearest point
func (s *Service) findNearestPointIndex(from model.Location, points []model.Location) int {
	nearestIdx := 0
	minDistance := -1.0
	
	for i, point := range points {
		distance := s.locationService.CalculateDistance(
			from.Latitude, from.Longitude,
			point.Latitude, point.Longitude,
		)
		
		if minDistance < 0 || distance < minDistance {
			minDistance = distance
			nearestIdx = i
		}
	}
	
	return nearestIdx
}

// calculateDuration estimates travel duration based on distance and mode
func (s *Service) calculateDuration(distanceKm float64, mode TransportMode) int {
	// Rough estimates of average travel speeds
	var speedKmh float64
	
	switch mode {
	case TransportModeDriving:
		// In Sierra Leone, roads are often in poor condition, so use conservative speed
		speedKmh = 30.0 // 30 km/h average for driving
	case TransportModeWalking:
		speedKmh = 5.0 // 5 km/h average for walking
	case TransportModeBicycling:
		speedKmh = 10.0 // 10 km/h average for bicycling
	default:
		speedKmh = 30.0 // Default to driving
	}
	
	// Calculate duration in seconds
	durationHours := distanceKm / speedKmh
	durationSeconds := int(durationHours * 3600)
	
	// Add some buffer for stops, traffic, etc.
	durationSeconds = int(float64(durationSeconds) * 1.2)
	
	return durationSeconds
}

// DescribeRoute creates a human-readable description of a route
func (s *Service) DescribeRoute(route *Route) []string {
	if route == nil || len(route.Points) < 2 {
		return []string{"Invalid route"}
	}
	
	descriptions := make([]string, 0, len(route.Points)-1)
	
	// Format each leg of the journey
	for i := 0; i < len(route.Points)-1; i++ {
		from := route.Points[i]
		to := route.Points[i+1]
		
		// Create description
		desc := ""
		
		// Add names if available
		if from.Name != "" && to.Name != "" {
			desc = "From " + from.Name + " to " + to.Name
		} else {
			desc = "From point " + PointToString(from.Location) + " to " + PointToString(to.Location)
		}
		
		// Add distance and duration
		desc += " (" + s.locationService.FormatDistance(from.DistanceToNext) + ", "
		desc += FormatDuration(from.Duration) + ")"
		
		descriptions = append(descriptions, desc)
	}
	
	// Add summary
	summary := "Total journey: " + s.locationService.FormatDistance(route.TotalDistance) + 
		", " + FormatDuration(route.TotalDuration)
	
	descriptions = append(descriptions, summary)
	
	return descriptions
}

// PointToString converts a location to a string
func PointToString(location model.Location) string {
	return fmt.Sprintf("%.6f, %.6f", location.Latitude, location.Longitude)
}

// FormatDuration formats duration in seconds to a human-readable string
func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%d sec", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%d min", seconds/60)
	} else {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}
