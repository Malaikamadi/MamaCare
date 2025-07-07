package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/visit/assignment"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// AssignCHWRequest is the request for assigning a CHW to a visit
type AssignCHWRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
	CHWID   string `json:"chw_id" validate:"required,uuid"`
}

// UnassignCHWRequest is the request for unassigning a CHW from a visit
type UnassignCHWRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
}

// GetCHWWorkloadRequest is the request for getting a CHW's workload
type GetCHWWorkloadRequest struct {
	CHWID     string `json:"chw_id" validate:"required,uuid"`
	StartDate string `json:"start_date" validate:"required,rfc3339"`
	EndDate   string `json:"end_date" validate:"required,rfc3339"`
}

// AssignVisitsByCatchmentAreaRequest is the request for assigning visits by catchment area
type AssignVisitsByCatchmentAreaRequest struct {
	FacilityID string `json:"facility_id" validate:"required,uuid"`
	StartDate  string `json:"start_date" validate:"required,rfc3339"`
	EndDate    string `json:"end_date" validate:"required,rfc3339"`
}

// OptimizeCHWRoutesRequest is the request for optimizing CHW routes
type OptimizeCHWRoutesRequest struct {
	CHWID string `json:"chw_id" validate:"required,uuid"`
	Date  string `json:"date" validate:"required,rfc3339"`
}

// BalanceWorkloadRequest is the request for balancing CHW workload
type BalanceWorkloadRequest struct {
	FacilityID string `json:"facility_id" validate:"required,uuid"`
	Date       string `json:"date" validate:"required,rfc3339"`
}

// UpdateVisitOrderRequest is the request for updating visit order
type UpdateVisitOrderRequest struct {
	CHWID    string   `json:"chw_id" validate:"required,uuid"`
	VisitIDs []string `json:"visit_ids" validate:"required,dive,uuid"`
}

// AssignmentHandler handles CHW assignment requests
type AssignmentHandler struct {
	hasura.BaseActionHandler
	assignmentService *assignment.Service
	validator         *validation.Validator
	log               logger.Logger
}

// NewAssignmentHandler creates a new assignment handler
func NewAssignmentHandler(
	log logger.Logger,
	assignmentService *assignment.Service,
	validator *validation.Validator,
) *AssignmentHandler {
	return &AssignmentHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		assignmentService: assignmentService,
		validator:         validator,
		log:               log,
	}
}

// AssignCHW assigns a CHW to a visit
func (h *AssignmentHandler) AssignCHW(w http.ResponseWriter, r *http.Request) {
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

	// Parse UUIDs
	visitID, err := uuid.Parse(req.VisitID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid visit ID"))
		return
	}

	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	// Assign CHW
	visit, err := h.assignmentService.AssignCHW(ctx, visitID, chwID)
	if err != nil {
		h.log.Error("Failed to assign CHW", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("CHW assigned successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
		"chw_id":     req.CHWID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// UnassignCHW unassigns a CHW from a visit
func (h *AssignmentHandler) UnassignCHW(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req UnassignCHWRequest
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

	// Parse UUID
	visitID, err := uuid.Parse(req.VisitID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid visit ID"))
		return
	}

	// Unassign CHW
	visit, err := h.assignmentService.UnassignCHW(ctx, visitID)
	if err != nil {
		h.log.Error("Failed to unassign CHW", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("CHW unassigned successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// GetCHWWorkload gets the workload for a CHW
func (h *AssignmentHandler) GetCHWWorkload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetCHWWorkloadRequest
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

	// Parse UUID
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid end date format"))
		return
	}

	// Get workload
	visits, err := h.assignmentService.GetCHWWorkload(ctx, chwID, startDate, endDate)
	if err != nil {
		h.log.Error("Failed to get CHW workload", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got CHW workload", logger.Fields{
		"request_id": reqID,
		"chw_id":     req.CHWID,
		"count":      len(visits),
	})

	response.WriteJSONResponse(w, reqID, visits)
}

// AssignVisitsByCatchmentArea assigns visits based on catchment areas
func (h *AssignmentHandler) AssignVisitsByCatchmentArea(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req AssignVisitsByCatchmentAreaRequest
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

	// Parse UUID
	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid facility ID"))
		return
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid start date format"))
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid end date format"))
		return
	}

	// Assign visits
	assignedCount, err := h.assignmentService.AssignVisitsByCatchmentArea(ctx, facilityID, startDate, endDate)
	if err != nil {
		h.log.Error("Failed to assign visits by catchment area", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Assigned visits by catchment area", logger.Fields{
		"request_id":     reqID,
		"facility_id":    req.FacilityID,
		"assigned_count": assignedCount,
	})

	response.WriteJSONResponse(w, reqID, struct {
		AssignedCount int `json:"assigned_count"`
	}{
		AssignedCount: assignedCount,
	})
}

// OptimizeCHWRoutes optimizes routes for a CHW
func (h *AssignmentHandler) OptimizeCHWRoutes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req OptimizeCHWRoutesRequest
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

	// Parse UUID
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid date format"))
		return
	}

	// Optimize routes
	optimizedRoute, err := h.assignmentService.OptimizeCHWRoutes(ctx, chwID, date)
	if err != nil {
		h.log.Error("Failed to optimize CHW routes", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Optimized CHW routes", logger.Fields{
		"request_id": reqID,
		"chw_id":     req.CHWID,
		"visit_count": len(optimizedRoute.Visits),
	})

	response.WriteJSONResponse(w, reqID, optimizedRoute)
}

// BalanceWorkload balances workload across CHWs
func (h *AssignmentHandler) BalanceWorkload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req BalanceWorkloadRequest
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

	// Parse UUID
	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid facility ID"))
		return
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid date format"))
		return
	}

	// Balance workload
	reassignedCount, err := h.assignmentService.BalanceWorkload(ctx, facilityID, date)
	if err != nil {
		h.log.Error("Failed to balance workload", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Balanced workload", logger.Fields{
		"request_id":       reqID,
		"facility_id":      req.FacilityID,
		"reassigned_count": reassignedCount,
	})

	response.WriteJSONResponse(w, reqID, struct {
		ReassignedCount int `json:"reassigned_count"`
	}{
		ReassignedCount: reassignedCount,
	})
}

// UpdateVisitOrder updates the order of visits for a CHW
func (h *AssignmentHandler) UpdateVisitOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req UpdateVisitOrderRequest
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

	// Parse UUIDs
	chwID, err := uuid.Parse(req.CHWID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid CHW ID"))
		return
	}

	visitIDs := make([]uuid.UUID, len(req.VisitIDs))
	for i, idStr := range req.VisitIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid visit ID: "+idStr))
			return
		}
		visitIDs[i] = id
	}

	// Update order
	if err := h.assignmentService.UpdateVisitOrder(ctx, chwID, visitIDs); err != nil {
		h.log.Error("Failed to update visit order", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"chw_id":     req.CHWID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Updated visit order", logger.Fields{
		"request_id": reqID,
		"chw_id":     req.CHWID,
		"visit_count": len(req.VisitIDs),
	})

	response.WriteJSONResponse(w, reqID, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}
