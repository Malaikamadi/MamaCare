package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/visit/scheduler"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// ScheduleVisitRequest is the request for scheduling a visit
type ScheduleVisitRequest struct {
	MotherID      string    `json:"mother_id" validate:"required,uuid"`
	FacilityID    string    `json:"facility_id" validate:"required,uuid"`
	ScheduledTime string    `json:"scheduled_time" validate:"required,rfc3339"`
	VisitType     string    `json:"visit_type" validate:"required,oneof=routine emergency follow_up"`
	Notes         string    `json:"notes,omitempty"`
}

// RescheduleVisitRequest is the request for rescheduling a visit
type RescheduleVisitRequest struct {
	VisitID          string `json:"visit_id" validate:"required,uuid"`
	NewScheduledTime string `json:"new_scheduled_time" validate:"required,rfc3339"`
}

// CancelVisitRequest is the request for cancelling a visit
type CancelVisitRequest struct {
	VisitID string `json:"visit_id" validate:"required,uuid"`
}

// GetUpcomingVisitsRequest is the request for getting upcoming visits
type GetUpcomingVisitsRequest struct {
	MotherID string `json:"mother_id" validate:"required,uuid"`
	Limit    int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
}

// GetVisitHistoryRequest is the request for getting visit history
type GetVisitHistoryRequest struct {
	MotherID string `json:"mother_id" validate:"required,uuid"`
	Limit    int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset   int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// GenerateAutomaticVisitsRequest is the request for generating automatic visits
type GenerateAutomaticVisitsRequest struct {
	MotherID   string `json:"mother_id" validate:"required,uuid"`
	FacilityID string `json:"facility_id" validate:"required,uuid"`
}

// GetVisitsByFacilityRequest is the request for getting visits by facility
type GetVisitsByFacilityRequest struct {
	FacilityID string `json:"facility_id" validate:"required,uuid"`
	StartDate  string `json:"start_date" validate:"required,rfc3339"`
	EndDate    string `json:"end_date" validate:"required,rfc3339"`
	Status     string `json:"status,omitempty" validate:"omitempty,oneof=scheduled checked_in completed cancelled"`
	Limit      int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset     int    `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// FindAvailableSlotsRequest is the request for finding available slots
type FindAvailableSlotsRequest struct {
	FacilityID      string `json:"facility_id" validate:"required,uuid"`
	Date            string `json:"date" validate:"required,rfc3339"`
	DurationMinutes int    `json:"duration_minutes,omitempty" validate:"omitempty,min=15,max=120"`
}

// SchedulerHandler handles visit scheduling requests
type SchedulerHandler struct {
	hasura.BaseActionHandler
	schedulerService *scheduler.Service
	validator        *validation.Validator
	log              logger.Logger
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(
	log logger.Logger,
	schedulerService *scheduler.Service,
	validator *validation.Validator,
) *SchedulerHandler {
	return &SchedulerHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		schedulerService:  schedulerService,
		validator:         validator,
		log:               log,
	}
}

// ScheduleVisit schedules a new visit
func (h *SchedulerHandler) ScheduleVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req ScheduleVisitRequest
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
	motherID, err := uuid.Parse(req.MotherID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid mother ID"))
		return
	}

	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid facility ID"))
		return
	}

	// Parse time
	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledTime)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid scheduled time format"))
		return
	}

	// Parse visit type
	visitType := model.VisitType(req.VisitType)

	// Schedule visit
	visit, err := h.schedulerService.ScheduleVisit(ctx, motherID, facilityID, scheduledTime, visitType, req.Notes)
	if err != nil {
		h.log.Error("Failed to schedule visit", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"mother_id":   req.MotherID,
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Visit scheduled successfully", logger.Fields{
		"request_id":  reqID,
		"visit_id":    visit.ID.String(),
		"mother_id":   req.MotherID,
		"facility_id": req.FacilityID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// RescheduleVisit reschedules an existing visit
func (h *SchedulerHandler) RescheduleVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req RescheduleVisitRequest
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

	// Parse time
	newScheduledTime, err := time.Parse(time.RFC3339, req.NewScheduledTime)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid scheduled time format"))
		return
	}

	// Reschedule visit
	visit, err := h.schedulerService.RescheduleVisit(ctx, visitID, newScheduledTime)
	if err != nil {
		h.log.Error("Failed to reschedule visit", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Visit rescheduled successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// CancelVisit cancels a visit
func (h *SchedulerHandler) CancelVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CancelVisitRequest
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

	// Cancel visit
	visit, err := h.schedulerService.CancelVisit(ctx, visitID)
	if err != nil {
		h.log.Error("Failed to cancel visit", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"visit_id":   req.VisitID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Visit cancelled successfully", logger.Fields{
		"request_id": reqID,
		"visit_id":   req.VisitID,
	})

	response.WriteJSONResponse(w, reqID, visit)
}

// GetUpcomingVisits gets upcoming visits for a mother
func (h *SchedulerHandler) GetUpcomingVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetUpcomingVisitsRequest
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
	motherID, err := uuid.Parse(req.MotherID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid mother ID"))
		return
	}

	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Get upcoming visits
	visits, err := h.schedulerService.GetUpcomingVisits(ctx, motherID, limit)
	if err != nil {
		h.log.Error("Failed to get upcoming visits", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"mother_id":  req.MotherID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got upcoming visits", logger.Fields{
		"request_id": reqID,
		"mother_id":  req.MotherID,
		"count":      len(visits),
	})

	response.WriteJSONResponse(w, reqID, visits)
}

// GetVisitHistory gets visit history for a mother
func (h *SchedulerHandler) GetVisitHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetVisitHistoryRequest
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
	motherID, err := uuid.Parse(req.MotherID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid mother ID"))
		return
	}

	// Set default values if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := req.Offset

	// Get visit history
	visits, err := h.schedulerService.GetVisitHistory(ctx, motherID, limit, offset)
	if err != nil {
		h.log.Error("Failed to get visit history", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"mother_id":  req.MotherID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got visit history", logger.Fields{
		"request_id": reqID,
		"mother_id":  req.MotherID,
		"count":      len(visits),
	})

	response.WriteJSONResponse(w, reqID, visits)
}

// GenerateAutomaticVisits generates automatic visits based on pregnancy stage
func (h *SchedulerHandler) GenerateAutomaticVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GenerateAutomaticVisitsRequest
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
	motherID, err := uuid.Parse(req.MotherID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid mother ID"))
		return
	}

	facilityID, err := uuid.Parse(req.FacilityID)
	if err != nil {
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid facility ID"))
		return
	}

	// Generate visits
	visits, err := h.schedulerService.GenerateAutomaticVisits(ctx, motherID, facilityID)
	if err != nil {
		h.log.Error("Failed to generate automatic visits", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"mother_id":   req.MotherID,
			"facility_id": req.FacilityID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Generated automatic visits", logger.Fields{
		"request_id":  reqID,
		"mother_id":   req.MotherID,
		"facility_id": req.FacilityID,
		"count":       len(visits),
	})

	response.WriteJSONResponse(w, reqID, visits)
}

// FindAvailableSlots finds available time slots for a facility
func (h *SchedulerHandler) FindAvailableSlots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req FindAvailableSlotsRequest
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

	// Set default duration if not provided
	durationMinutes := req.DurationMinutes
	if durationMinutes <= 0 {
		durationMinutes = 30
	}

	// Find available slots
	slots, err := h.schedulerService.FindAvailableSlots(ctx, facilityID, date, durationMinutes)
	if err != nil {
		h.log.Error("Failed to find available slots", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"facility_id": req.FacilityID,
			"date":        req.Date,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Found available slots", logger.Fields{
		"request_id":  reqID,
		"facility_id": req.FacilityID,
		"date":        req.Date,
		"count":       len(slots),
	})

	response.WriteJSONResponse(w, reqID, slots)
}
