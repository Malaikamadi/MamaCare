package action

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/geo/territory"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// CreateTerritoryRequest is the request for creating a territory
type CreateTerritoryRequest struct {
	CHWID       string           `json:"chw_id" validate:"required,uuid"`
	Name        string           `json:"name" validate:"required,min=3,max=100"`
	District    string           `json:"district" validate:"required,min=3,max=100"`
	Description string           `json:"description,omitempty"`
	Boundaries  []model.Location `json:"boundaries" validate:"required,min=3"`
}

// TerritoryHandler handles territory-related requests
type TerritoryHandler struct {
	hasura.BaseActionHandler
	territoryService *territory.Service
	validator        *validation.Validator
	log              logger.Logger
}

// NewTerritoryHandler creates a new territory handler
func NewTerritoryHandler(
	log logger.Logger,
	territoryService *territory.Service,
	validator *validation.Validator,
) *TerritoryHandler {
	return &TerritoryHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		territoryService:  territoryService,
		validator:         validator,
		log:               log,
	}
}

// CreateTerritory creates a new territory for a CHW
func (h *TerritoryHandler) CreateTerritory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CreateTerritoryRequest
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

	// Parse CHW ID
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		h.log.Error("Invalid CHW ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	// Create territory
	territory, err := h.territoryService.CreateTerritory(
		ctx,
		chwID,
		req.Name,
		req.District,
		req.Description,
		req.Boundaries,
	)
	if err != nil {
		h.log.Error("Failed to create territory", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
			"name":       req.Name,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Created territory", logger.Fields{
		"request_id":   reqID,
		"territory_id": territory.ID.String(),
		"chw_id":       req.CHWID,
		"name":         req.Name,
	})

	response.WriteJSONResponse(w, reqID, territory)
}

// GetTerritoryRequest is the request for getting a territory
type GetTerritoryRequest struct {
	TerritoryID string `json:"territory_id" validate:"required,uuid"`
}

// GetTerritory gets a territory by ID
func (h *TerritoryHandler) GetTerritory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetTerritoryRequest
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

	// Parse territory ID
	territoryID, err := uuid.Parse(req.TerritoryID)
	if err != nil {
		h.log.Error("Invalid territory ID", logger.Fields{
			"request_id":   reqID,
			"error":        err.Error(),
			"territory_id": req.TerritoryID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid territory ID"))
		return
	}

	// Get territory
	territory, err := h.territoryService.GetTerritory(ctx, territoryID)
	if err != nil {
		h.log.Error("Failed to get territory", logger.Fields{
			"request_id":   reqID,
			"error":        err.Error(),
			"territory_id": req.TerritoryID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got territory", logger.Fields{
		"request_id":   reqID,
		"territory_id": req.TerritoryID,
	})

	response.WriteJSONResponse(w, reqID, territory)
}

// GetTerritoryForCHWRequest is the request for getting a territory for a CHW
type GetTerritoryForCHWRequest struct {
	CHWID string `json:"chw_id" validate:"required,uuid"`
}

// GetTerritoryForCHW gets a territory for a CHW
func (h *TerritoryHandler) GetTerritoryForCHW(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetTerritoryForCHWRequest
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

	// Parse CHW ID
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		h.log.Error("Invalid CHW ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	// Get territory for CHW
	territory, err := h.territoryService.GetTerritoryForCHW(ctx, chwID)
	if err != nil {
		h.log.Error("Failed to get territory for CHW", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got territory for CHW", logger.Fields{
		"request_id":   reqID,
		"chw_id":       req.CHWID,
		"territory_id": territory.ID.String(),
	})

	response.WriteJSONResponse(w, reqID, territory)
}

// FindTerritoryForLocationRequest is the request for finding a territory for a location
type FindTerritoryForLocationRequest struct {
	Latitude  float64 `json:"latitude" validate:"required,latitude"`
	Longitude float64 `json:"longitude" validate:"required,longitude"`
}

// FindTerritoryForLocation finds the territory that contains a location
func (h *TerritoryHandler) FindTerritoryForLocation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req FindTerritoryForLocationRequest
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

	// Find territory for location
	territory, err := h.territoryService.FindTerritoryForLocation(ctx, req.Latitude, req.Longitude)
	if err != nil {
		h.log.Error("Failed to find territory for location", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"lat":        req.Latitude,
			"lng":        req.Longitude,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Found territory for location", logger.Fields{
		"request_id":   reqID,
		"lat":          req.Latitude,
		"lng":          req.Longitude,
		"territory_id": territory.ID.String(),
	})

	response.WriteJSONResponse(w, reqID, territory)
}

// AssignCHWRequest is the request for assigning a CHW to a territory
type AssignCHWRequest struct {
	CHWID      string `json:"chw_id" validate:"required,uuid"`
	TerritoryID string `json:"territory_id" validate:"required,uuid"`
}

// AssignCHWToTerritory assigns a CHW to a territory
func (h *TerritoryHandler) AssignCHWToTerritory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req AssignCHWRequest
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

	// Parse CHW ID and territory ID
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		h.log.Error("Invalid CHW ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	territoryID, err := uuid.Parse(req.TerritoryID)
	if err != nil {
		h.log.Error("Invalid territory ID", logger.Fields{
			"request_id":   reqID,
			"error":        err.Error(),
			"territory_id": req.TerritoryID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid territory ID"))
		return
	}

	// Assign CHW to territory
	result, err := h.territoryService.AssignCHWToTerritory(ctx, chwID, territoryID)
	if err != nil {
		h.log.Error("Failed to assign CHW to territory", logger.Fields{
			"request_id":   reqID,
			"error":        err.Error(),
			"chw_id":       req.CHWID,
			"territory_id": req.TerritoryID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Assigned CHW to territory", logger.Fields{
		"request_id":   reqID,
		"chw_id":       req.CHWID,
		"territory_id": req.TerritoryID,
		"mother_count": result.MotherCount,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// GetMothersInTerritoryRequest is the request for getting mothers in a territory
type GetMothersInTerritoryRequest struct {
	TerritoryID string `json:"territory_id" validate:"required,uuid"`
}

// GetMothersInTerritory gets all mothers in a territory
func (h *TerritoryHandler) GetMothersInTerritory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetMothersInTerritoryRequest
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

	// Parse territory ID
	territoryID, err := uuid.Parse(req.TerritoryID)
	if err != nil {
		h.log.Error("Invalid territory ID", logger.Fields{
			"request_id":   reqID,
			"error":        err.Error(),
			"territory_id": req.TerritoryID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid territory ID"))
		return
	}

	// Get mothers in territory
	mothers, err := h.territoryService.GetMothersInTerritory(ctx, territoryID)
	if err != nil {
		h.log.Error("Failed to get mothers in territory", logger.Fields{
			"request_id":   reqID,
			"error":        err.Error(),
			"territory_id": req.TerritoryID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got mothers in territory", logger.Fields{
		"request_id":   reqID,
		"territory_id": req.TerritoryID,
		"mother_count": len(mothers),
	})

	response.WriteJSONResponse(w, reqID, mothers)
}
