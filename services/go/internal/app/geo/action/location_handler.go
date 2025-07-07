package action

import (
	"net/http"

	"github.com/mamacare/services/internal/app/geo/location"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/logger"
)

// ValidateLocationRequest is the request for validating a location
type ValidateLocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

// LocationHandler handles location-related requests
type LocationHandler struct {
	hasura.BaseActionHandler
	locationService *location.Service
	validator       *validation.Validator
	log             logger.Logger
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(
	log logger.Logger,
	locationService *location.Service,
	validator *validation.Validator,
) *LocationHandler {
	return &LocationHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		locationService:   locationService,
		validator:         validator,
		log:               log,
	}
}

// ValidateLocation validates a location's coordinates
func (h *LocationHandler) ValidateLocation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req ValidateLocationRequest
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

	// Validate coordinates
	location, err := h.locationService.ValidateCoordinates(ctx, req.Latitude, req.Longitude)
	if err != nil {
		h.log.Error("Failed to validate coordinates", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"lat":        req.Latitude,
			"lng":        req.Longitude,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Validated coordinates", logger.Fields{
		"request_id": reqID,
		"lat":        req.Latitude,
		"lng":        req.Longitude,
	})

	response.WriteJSONResponse(w, reqID, location)
}

// NormalizeAddressRequest is the request for normalizing an address
type NormalizeAddressRequest struct {
	Address string `json:"address" validate:"required,min=5,max=500"`
}

// NormalizeAddress normalizes an address and returns its coordinates
func (h *LocationHandler) NormalizeAddress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req NormalizeAddressRequest
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

	// Normalize address
	normalizedLocation, err := h.locationService.NormalizeAddress(ctx, req.Address)
	if err != nil {
		h.log.Error("Failed to normalize address", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"address":    req.Address,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Normalized address", logger.Fields{
		"request_id": reqID,
		"address":    req.Address,
		"lat":        normalizedLocation.Latitude,
		"lng":        normalizedLocation.Longitude,
	})

	response.WriteJSONResponse(w, reqID, normalizedLocation)
}

// CalculateDistanceRequest is the request for calculating distance between two points
type CalculateDistanceRequest struct {
	Point1 struct {
		Latitude  float64 `json:"latitude" validate:"required,latitude"`
		Longitude float64 `json:"longitude" validate:"required,longitude"`
	} `json:"point1" validate:"required"`
	Point2 struct {
		Latitude  float64 `json:"latitude" validate:"required,latitude"`
		Longitude float64 `json:"longitude" validate:"required,longitude"`
	} `json:"point2" validate:"required"`
}

// CalculateDistance calculates the distance between two points
func (h *LocationHandler) CalculateDistance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CalculateDistanceRequest
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

	// Convert to model.Location
	point1 := model.Location{
		Latitude:  req.Point1.Latitude,
		Longitude: req.Point1.Longitude,
	}
	point2 := model.Location{
		Latitude:  req.Point2.Latitude,
		Longitude: req.Point2.Longitude,
	}

	// Calculate distance
	distance, err := h.locationService.CalculateDistance(ctx, point1, point2)
	if err != nil {
		h.log.Error("Failed to calculate distance", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"lat1":       req.Point1.Latitude,
			"lng1":       req.Point1.Longitude,
			"lat2":       req.Point2.Latitude,
			"lng2":       req.Point2.Longitude,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Prepare response
	result := struct {
		DistanceKm     float64 `json:"distance_km"`
		DistanceMeters float64 `json:"distance_meters"`
		DistanceMiles  float64 `json:"distance_miles"`
	}{
		DistanceKm:     distance,
		DistanceMeters: distance * 1000,
		DistanceMiles:  distance * 0.621371,
	}

	h.log.Info("Calculated distance", logger.Fields{
		"request_id": reqID,
		"distance":   distance,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// GetLocationInfoRequest is the request for getting information about a location
type GetLocationInfoRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

// GetLocationInfo gets information about a location
func (h *LocationHandler) GetLocationInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetLocationInfoRequest
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

	// Get location info
	locationInfo, err := h.locationService.GetLocationInfo(ctx, req.Latitude, req.Longitude)
	if err != nil {
		h.log.Error("Failed to get location info", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"lat":        req.Latitude,
			"lng":        req.Longitude,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got location info", logger.Fields{
		"request_id": reqID,
		"lat":        req.Latitude,
		"lng":        req.Longitude,
	})

	response.WriteJSONResponse(w, reqID, locationInfo)
}
