package action

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/geo/facility"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// FindNearbyFacilitiesRequest is the request for finding nearby facilities
type FindNearbyFacilitiesRequest struct {
	Latitude       float64               `json:"latitude" validate:"required,latitude"`
	Longitude      float64               `json:"longitude" validate:"required,longitude"`
	RadiusKm       float64               `json:"radius_km,omitempty"`
	FacilityTypes  []model.FacilityType  `json:"facility_types,omitempty"`
	Services       []string              `json:"services,omitempty"`
	MaxDistance    float64               `json:"max_distance,omitempty"`
	OpenNow        bool                  `json:"open_now,omitempty"`
	MinCapacity    int                   `json:"min_capacity,omitempty"`
	District       string                `json:"district,omitempty"`
}

// FacilityHandler handles facility-related requests
type FacilityHandler struct {
	hasura.BaseActionHandler
	facilityService *facility.Service
	validator       *validation.Validator
	log             logger.Logger
}

// NewFacilityHandler creates a new facility handler
func NewFacilityHandler(
	log logger.Logger,
	facilityService *facility.Service,
	validator *validation.Validator,
) *FacilityHandler {
	return &FacilityHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		facilityService:   facilityService,
		validator:         validator,
		log:               log,
	}
}

// FindNearbyFacilities finds healthcare facilities near a location
func (h *FacilityHandler) FindNearbyFacilities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req FindNearbyFacilitiesRequest
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

	// Create filter
	filter := &facility.FacilityFilter{
		Types:       req.FacilityTypes,
		Services:    req.Services,
		MaxDistance: req.MaxDistance,
		OpenNow:     req.OpenNow,
		MinCapacity: req.MinCapacity,
		District:    req.District,
	}

	// Default radius if not provided
	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 10.0 // Default 10km radius
	}

	// Find nearby facilities
	facilities, err := h.facilityService.FindNearbyFacilities(
		ctx,
		req.Latitude, req.Longitude,
		radiusKm,
		filter,
	)
	if err != nil {
		h.log.Error("Failed to find nearby facilities", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"lat":        req.Latitude,
			"lng":        req.Longitude,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Found nearby facilities", logger.Fields{
		"request_id":      reqID,
		"lat":             req.Latitude,
		"lng":             req.Longitude,
		"facilities_count": len(facilities),
	})

	response.WriteJSONResponse(w, reqID, facilities)
}

// SearchFacilitiesRequest is the request for searching facilities
type SearchFacilitiesRequest struct {
	Query           string                `json:"query" validate:"required,min=1"`
	FacilityTypes   []model.FacilityType  `json:"facility_types,omitempty"`
	Services        []string              `json:"services,omitempty"`
	OpenNow         bool                  `json:"open_now,omitempty"`
	MinCapacity     int                   `json:"min_capacity,omitempty"`
	District        string                `json:"district,omitempty"`
}

// SearchFacilities searches facilities by query
func (h *FacilityHandler) SearchFacilities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req SearchFacilitiesRequest
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

	// Create filter
	filter := &facility.FacilityFilter{
		Types:       req.FacilityTypes,
		Services:    req.Services,
		OpenNow:     req.OpenNow,
		MinCapacity: req.MinCapacity,
		District:    req.District,
	}

	// Search facilities
	facilities, err := h.facilityService.SearchFacilities(
		ctx,
		req.Query,
		filter,
	)
	if err != nil {
		h.log.Error("Failed to search facilities", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"query":      req.Query,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Searched facilities", logger.Fields{
		"request_id":      reqID,
		"query":           req.Query,
		"facilities_count": len(facilities),
	})

	response.WriteJSONResponse(w, reqID, facilities)
}

// GetFacilityRequest is the request for getting a facility
type GetFacilityRequest struct {
	FacilityID string `json:"facility_id" validate:"required,uuid"`
}

// GetFacility gets a facility by ID
func (h *FacilityHandler) GetFacility(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetFacilityRequest
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

	// Parse facility ID
	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		h.log.Error("Invalid facility ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid facility ID"))
		return
	}

	// Get facility
	facility, err := h.facilityService.FindFacilityByID(ctx, facilityID)
	if err != nil {
		h.log.Error("Failed to get facility", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got facility", logger.Fields{
		"request_id": reqID,
		"facility_id": req.FacilityID,
	})

	response.WriteJSONResponse(w, reqID, facility)
}
