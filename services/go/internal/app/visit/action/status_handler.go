package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/visit/status"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// CheckInVisitRequest is the request for checking in a visit
type CheckInVisitRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
}

// CompleteVisitRequest is the request for completing a visit
type CompleteVisitRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
	Notes   string `json:"notes,omitempty"`
}

// GetVisitStatusRequest is the request for getting visit status
type GetVisitStatusRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
}

// GetOverdueVisitsRequest is the request for getting overdue visits
type GetOverdueVisitsRequest struct {
	Limit  int `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// ScheduleFollowUpRequest is the request for scheduling a follow-up visit
type ScheduleFollowUpRequest struct {
	OriginalVisitID string `json:"original_visit_id" validate:"required,uuid"`
	ScheduledTime   string `json:"scheduled_time" validate:"required,rfc3339"`
	Notes           string `json:"notes,omitempty"`
}

// UpdateVisitNotesRequest is the request for updating visit notes
type UpdateVisitNotesRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
	Notes   string `json:"notes" validate:"required"`
}

// StatusHandler handles visit status requests
type StatusHandler struct {
	hasura.BaseActionHandler
	statusService *status.Service
	validator     *validation.Validator
	log           logger.Logger
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(
	log logger.Logger,
	statusService *status.Service,
	validator *validation.Validator,
) *StatusHandler {
	return &StatusHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		statusService:     statusService,
		validator:         validator,
		log:               log,
	}
}

// CheckInVisit checks in a visit
func (h *StatusHandler) CheckInVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CheckInVisitRequest
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

	// Check in visit
	visit, err := h.statusService.CheckInVisit(ctx, visitID)
	if err != nil {
		h.log.Error("Failed to check in visit", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Visit checked in successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// CompleteVisit completes a visit
func (h *StatusHandler) CompleteVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CompleteVisitRequest
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

	// Complete visit
	visit, err := h.statusService.CompleteVisit(ctx, visitID, req.Notes)
	if err != nil {
		h.log.Error("Failed to complete visit", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Visit completed successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// GetVisitStatus gets the status of a visit
func (h *StatusHandler) GetVisitStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetVisitStatusRequest
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

	// Get visit status
	statusDetails, err := h.statusService.GetVisitStatus(ctx, visitID)
	if err != nil {
		h.log.Error("Failed to get visit status", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got visit status", logger.Fields{
		"request_id":   reqID,
		"visit_id":     req.VisitID,
		"visit_status": string(statusDetails.Visit.Status),
	})

	response.WriteJSONResponse(w, reqID, statusDetails)
}

// GetOverdueVisits gets overdue visits
func (h *StatusHandler) GetOverdueVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetOverdueVisitsRequest
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

	// Set default values if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset

	// Get overdue visits
	visits, err := h.statusService.GetOverdueVisits(ctx, limit, offset)
	if err != nil {
		h.log.Error("Failed to get overdue visits", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got overdue visits", logger.Fields{
		"request_id": reqID,
		"count":      len(visits),
	})

	response.WriteJSONResponse(w, reqID, visits)
}

// ScheduleFollowUp schedules a follow-up visit
func (h *StatusHandler) ScheduleFollowUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req ScheduleFollowUpRequest
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
	originalVisitID, err := uuid.Parse(req.OriginalVisitID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid original visit ID"))
		return
	}

	// Parse time
	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledTime)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid scheduled time format"))
		return
	}

	// Schedule follow-up
	visit, err := h.statusService.ScheduleFollowUp(ctx, originalVisitID, scheduledTime, req.Notes)
	if err != nil {
		h.log.Error("Failed to schedule follow-up visit", logger.Fields{
			"request_id":        reqID,
			"error":             err.Error(),
			"original_visit_id": req.OriginalVisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Follow-up visit scheduled successfully", logger.Fields{
		"request_id":        reqID,
		"visit_id":          visit.ID.String(),
		"original_visit_id": req.OriginalVisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// UpdateVisitNotes updates the notes for a visit
func (h *StatusHandler) UpdateVisitNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req UpdateVisitNotesRequest
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

	// Update notes
	visit, err := h.statusService.UpdateVisitNotes(ctx, visitID, req.Notes)
	if err != nil {
		h.log.Error("Failed to update visit notes", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Visit notes updated successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// MarkMissedVisits marks missed visits
func (h *StatusHandler) MarkMissedVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	// This endpoint doesn't require a request body
	// It will find and mark all missed visits

	// Mark missed visits
	missedCount, err := h.statusService.MarkMissedVisits(ctx)
	if err != nil {
		h.log.Error("Failed to mark missed visits", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Marked missed visits", logger.Fields{
		"request_id":   reqID,
		"missed_count": missedCount,
	})

	response.WriteJSONResponse(w, reqID, struct {
		MissedCount int `json:"missed_count"`
	}{
		MissedCount: missedCount,
	})
}
