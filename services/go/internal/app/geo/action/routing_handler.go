package action

import (
	"net/http"
	"time"

	"github.com/mamacare/services/internal/app/geo/routing"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// RouteCalculationRequest is the request for calculating a route
type RouteCalculationRequest struct {
	Points []struct {
		Latitude  float64 `json:"latitude" validate:"required,latitude"`
		Longitude float64 `json:"longitude" validate:"required,longitude"`
	} `json:"points" validate:"required,min=2"`
	TransportMode string    `json:"transport_mode,omitempty"`
	AvoidHighways bool      `json:"avoid_highways,omitempty"`
	DepartureTime string    `json:"departure_time,omitempty"`
	Optimize      bool      `json:"optimize,omitempty"`
}

// RoutingHandler handles routing-related requests
type RoutingHandler struct {
	hasura.BaseActionHandler
	routingService *routing.Service
	validator      *validation.Validator
	log            logger.Logger
}

// NewRoutingHandler creates a new routing handler
func NewRoutingHandler(
	log logger.Logger,
	routingService *routing.Service,
	validator *validation.Validator,
) *RoutingHandler {
	return &RoutingHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		routingService:    routingService,
		validator:         validator,
		log:               log,
	}
}

// CalculateRoute calculates a route between multiple points
func (h *RoutingHandler) CalculateRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req RouteCalculationRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Convert points to model.Location
	points := make([]model.Location, len(req.Points))
	for i, p := range req.Points {
		points[i] = model.Location{
			Latitude:  p.Latitude,
			Longitude: p.Longitude,
		}
	}

	// Parse transport mode
	var transportMode routing.TransportMode
	switch req.TransportMode {
	case "walking":
		transportMode = routing.TransportModeWalking
	case "bicycling":
		transportMode = routing.TransportModeBicycling
	default:
		transportMode = routing.TransportModeDriving
	}

	// Parse departure time
	var departureTime time.Time
	if req.DepartureTime != "" {
		var err error
		departureTime, err = time.Parse(time.RFC3339, req.DepartureTime)
		if err != nil {
			h.log.Error("Invalid departure time format", logger.Fields{
				"request_id":     reqID,
				"error":          err.Error(),
				"departure_time": req.DepartureTime,
			})
			response.WriteErrorResponse(w, reqID, 
				errorx.New(errorx.BadRequest, "Invalid departure time format. Use RFC3339 format."))
			return
		}
	} else {
		departureTime = time.Now()
	}

	// Create options
	options := &routing.RouteOptions{
		TransportMode:  transportMode,
		AvoidHighways:  req.AvoidHighways,
		DepartureTime:  departureTime,
		Optimize:       req.Optimize,
	}

	// Calculate route
	route, err := h.routingService.CalculateRoute(ctx, points, options)
	if err != nil {
		h.log.Error("Failed to calculate route", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"points":     len(points),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Get route description
	descriptions := h.routingService.DescribeRoute(route)

	// Create response
	result := struct {
		Route        *routing.Route `json:"route"`
		Descriptions []string       `json:"descriptions"`
	}{
		Route:        route,
		Descriptions: descriptions,
	}

	h.log.Info("Calculated route", logger.Fields{
		"request_id":     reqID,
		"points":         len(points),
		"total_distance": route.TotalDistance,
		"total_duration": route.TotalDuration,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// VisitPlanningRequest is the request for planning CHW visits
type VisitPlanningRequest struct {
	CHWLocation struct {
		Latitude  float64 `json:"latitude" validate:"required,latitude"`
		Longitude float64 `json:"longitude" validate:"required,longitude"`
	} `json:"chw_location" validate:"required"`
	MotherLocations []struct {
		Latitude  float64 `json:"latitude" validate:"required,latitude"`
		Longitude float64 `json:"longitude" validate:"required,longitude"`
	} `json:"mother_locations" validate:"required,min=1"`
	TransportMode string `json:"transport_mode,omitempty"`
	DepartureTime string `json:"departure_time,omitempty"`
}

// PlanVisits plans an optimized route for CHW visits
func (h *RoutingHandler) PlanVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req VisitPlanningRequest
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate request
	if err := h.validator.Validate(req); err != nil {
		h.log.Error("Invalid request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Convert CHW location to model.Location
	chwLocation := model.Location{
		Latitude:  req.CHWLocation.Latitude,
		Longitude: req.CHWLocation.Longitude,
	}

	// Convert mother locations to model.Location
	motherLocations := make([]model.Location, len(req.MotherLocations))
	for i, ml := range req.MotherLocations {
		motherLocations[i] = model.Location{
			Latitude:  ml.Latitude,
			Longitude: ml.Longitude,
		}
	}

	// Parse transport mode
	var transportMode routing.TransportMode
	switch req.TransportMode {
	case "walking":
		transportMode = routing.TransportModeWalking
	case "bicycling":
		transportMode = routing.TransportModeBicycling
	default:
		transportMode = routing.TransportModeDriving
	}

	// Parse departure time
	var departureTime time.Time
	if req.DepartureTime != "" {
		var err error
		departureTime, err = time.Parse(time.RFC3339, req.DepartureTime)
		if err != nil {
			h.log.Error("Invalid departure time format", logger.Fields{
				"request_id":     reqID,
				"error":          err.Error(),
				"departure_time": req.DepartureTime,
			})
			response.WriteErrorResponse(w, reqID, 
				errorx.New(errorx.BadRequest, "Invalid departure time format. Use RFC3339 format."))
			return
		}
	} else {
		departureTime = time.Now()
	}

	// Create options
	options := &routing.RouteOptions{
		TransportMode:  transportMode,
		DepartureTime:  departureTime,
		Optimize:       true, // Always optimize visit planning
	}

	// Calculate visit route
	route, err := h.routingService.CalculateVisitRoute(ctx, chwLocation, motherLocations, options)
	if err != nil {
		h.log.Error("Failed to plan visits", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"mothers":     len(motherLocations),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Get route description
	descriptions := h.routingService.DescribeRoute(route)

	// Create response
	result := struct {
		Route        *routing.Route `json:"route"`
		Descriptions []string       `json:"descriptions"`
	}{
		Route:        route,
		Descriptions: descriptions,
	}

	h.log.Info("Planned CHW visits", logger.Fields{
		"request_id":     reqID,
		"mothers":        len(motherLocations),
		"total_distance": route.TotalDistance,
		"total_duration": route.TotalDuration,
	})

	response.WriteJSONResponse(w, reqID, result)
}
