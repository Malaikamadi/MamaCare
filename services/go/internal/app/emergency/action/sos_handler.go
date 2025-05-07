package action

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/emergency/sos"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
)

// SOSHandler handles all SOS event related actions
type SOSHandler struct {
	*hasura.BaseActionHandler
	sosService *sos.Service
	logger     logger.Logger
}

// ReportSOSRequest defines the request payload for reporting an SOS event
type ReportSOSRequest struct {
	MotherID    string  `json:"mother_id"`
	ReportedBy  string  `json:"reported_by"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Nature      string  `json:"nature"`
	Description string  `json:"description"`
}

// ReportSOSResponse defines the response payload for reporting an SOS event
type ReportSOSResponse struct {
	ID          string    `json:"id"`
	MotherID    string    `json:"mother_id"`
	ReportedBy  string    `json:"reported_by"`
	Status      string    `json:"status"`
	Nature      string    `json:"nature"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FacilityID  *string   `json:"facility_id,omitempty"`
	AmbulanceID *string   `json:"ambulance_id,omitempty"`
	Priority    int       `json:"priority"`
	ETA         *string   `json:"eta,omitempty"`
}

// UpdateSOSRequest defines the request payload for updating an SOS event status
type UpdateSOSRequest struct {
	SOSID  string `json:"sos_id"`
	Status string `json:"status"`
}

// GetSOSRequest defines the request payload for getting an SOS event
type GetSOSRequest struct {
	SOSID string `json:"sos_id"`
}

// SOSResponse defines the standard SOS event response
type SOSResponse struct {
	ID          string    `json:"id"`
	MotherID    string    `json:"mother_id"`
	ReportedBy  string    `json:"reported_by"`
	Status      string    `json:"status"`
	Nature      string    `json:"nature"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FacilityID  *string   `json:"facility_id,omitempty"`
	AmbulanceID *string   `json:"ambulance_id,omitempty"`
	Priority    int       `json:"priority"`
	ETA         *string   `json:"eta,omitempty"`
}

// AssignFacilityRequest defines the request payload for assigning a facility to an SOS event
type AssignFacilityRequest struct {
	SOSID      string `json:"sos_id"`
	FacilityID string `json:"facility_id"`
}

// GetSOSByMotherRequest defines the request payload for getting SOS events by mother ID
type GetSOSByMotherRequest struct {
	MotherID string `json:"mother_id"`
}

// SOSListResponse defines the response payload for lists of SOS events
type SOSListResponse struct {
	SOSEvents []SOSResponse `json:"sos_events"`
}

// NewSOSHandler creates a new SOS handler
func NewSOSHandler(sosService *sos.Service, logger logger.Logger) *SOSHandler {
	return &SOSHandler{
		BaseActionHandler: hasura.NewBaseActionHandler(logger),
		sosService: sosService,
		logger:     logger,
	}
}

// ReportSOS handles reporting a new SOS event
func (h *SOSHandler) ReportSOS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &ReportSOSRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse report SOS request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	sosReq := req.(*ReportSOSRequest)

	// Convert string IDs to UUID
	motherID, err := uuid.Parse(sosReq.MotherID)
	if err != nil {
		h.logger.Error(ctx, "Invalid mother ID", logger.FieldsMap{
			"error":      err.Error(),
			"mother_id":  sosReq.MotherID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid mother ID", requestID)
		return
	}

	reportedByID, err := uuid.Parse(sosReq.ReportedBy)
	if err != nil {
		h.logger.Error(ctx, "Invalid reporter ID", logger.FieldsMap{
			"error":      err.Error(),
			"reporter_id": sosReq.ReportedBy,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid reporter ID", requestID)
		return
	}

	// Convert nature string to enum
	var nature model.SOSEventNature
	switch sosReq.Nature {
	case "labor":
		nature = model.SOSEventNatureLabor
	case "bleeding":
		nature = model.SOSEventNatureBleeding
	case "accident":
		nature = model.SOSEventNatureAccident
	case "other":
		nature = model.SOSEventNatureOther
	default:
		h.logger.Error(ctx, "Invalid SOS event nature", logger.FieldsMap{
			"nature":     sosReq.Nature,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS event nature", requestID)
		return
	}

	// Call service
	sosEvent, err := h.sosService.ReportSOSEvent(
		ctx,
		motherID,
		reportedByID,
		sosReq.Latitude,
		sosReq.Longitude,
		nature,
		sosReq.Description,
	)
	if err != nil {
		h.logger.Error(ctx, "Failed to report SOS event", logger.FieldsMap{
			"error":      err.Error(),
			"mother_id":  sosReq.MotherID,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	var facilityIDStr *string
	if sosEvent.FacilityID != nil {
		idStr := sosEvent.FacilityID.String()
		facilityIDStr = &idStr
	}

	var ambulanceIDStr *string
	if sosEvent.AmbulanceID != nil {
		idStr := sosEvent.AmbulanceID.String()
		ambulanceIDStr = &idStr
	}

	var etaStr *string
	if sosEvent.ETA != nil {
		formattedETA := sosEvent.ETA.Format(time.RFC3339)
		etaStr = &formattedETA
	}

	resp := ReportSOSResponse{
		ID:          sosEvent.ID.String(),
		MotherID:    sosEvent.MotherID.String(),
		ReportedBy:  sosEvent.ReportedBy.String(),
		Status:      string(sosEvent.Status),
		Nature:      string(sosEvent.Nature),
		Description: sosEvent.Description,
		Latitude:    sosEvent.Location.Latitude,
		Longitude:   sosEvent.Location.Longitude,
		CreatedAt:   sosEvent.CreatedAt,
		UpdatedAt:   sosEvent.UpdatedAt,
		FacilityID:  facilityIDStr,
		AmbulanceID: ambulanceIDStr,
		Priority:    sosEvent.Priority,
		ETA:         etaStr,
	}

	h.logger.Info(ctx, "SOS event reported successfully", logger.FieldsMap{
		"sos_id":     sosEvent.ID.String(),
		"mother_id":  sosEvent.MotherID.String(),
		"request_id": requestID,
	})

	response.WriteJSONResponse(w, http.StatusCreated, resp, requestID)
}

// UpdateSOSStatus handles updating the status of an SOS event
func (h *SOSHandler) UpdateSOSStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &UpdateSOSRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse update SOS status request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	updateReq := req.(*UpdateSOSRequest)

	// Convert string ID to UUID
	sosID, err := uuid.Parse(updateReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     updateReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	// Convert status string to enum
	var status model.SOSEventStatus
	switch updateReq.Status {
	case "reported":
		status = model.SOSEventStatusReported
	case "dispatched":
		status = model.SOSEventStatusDispatched
	case "resolved":
		status = model.SOSEventStatusResolved
	case "cancelled":
		status = model.SOSEventStatusCancelled
	default:
		h.logger.Error(ctx, "Invalid SOS event status", logger.FieldsMap{
			"status":     updateReq.Status,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS event status", requestID)
		return
	}

	// Call service
	sosEvent, err := h.sosService.UpdateSOSEventStatus(ctx, sosID, status)
	if err != nil {
		h.logger.Error(ctx, "Failed to update SOS event status", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     updateReq.SOSID,
			"status":     updateReq.Status,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response using helper function
	resp := sosEventToResponse(sosEvent)

	h.logger.Info(ctx, "SOS event status updated successfully", logger.FieldsMap{
		"sos_id":     sosEvent.ID.String(),
		"status":     string(sosEvent.Status),
		"request_id": requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// GetSOS handles getting an SOS event by ID
func (h *SOSHandler) GetSOS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &GetSOSRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse get SOS request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	getReq := req.(*GetSOSRequest)

	// Convert string ID to UUID
	sosID, err := uuid.Parse(getReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     getReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	// Call service
	sosEvent, err := h.sosService.GetSOSEvent(ctx, sosID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get SOS event", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     getReq.SOSID,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := sosEventToResponse(sosEvent)

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// AssignFacility handles assigning a facility to an SOS event
func (h *SOSHandler) AssignFacility(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &AssignFacilityRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse assign facility request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	assignReq := req.(*AssignFacilityRequest)

	// Convert string IDs to UUID
	sosID, err := uuid.Parse(assignReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     assignReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	facilityID, err := uuid.Parse(assignReq.FacilityID)
	if err != nil {
		h.logger.Error(ctx, "Invalid facility ID", logger.FieldsMap{
			"error":       err.Error(),
			"facility_id": assignReq.FacilityID,
			"request_id":  requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid facility ID", requestID)
		return
	}

	// Call service
	sosEvent, err := h.sosService.AssignFacilityToSOSEvent(ctx, sosID, facilityID)
	if err != nil {
		h.logger.Error(ctx, "Failed to assign facility to SOS event", logger.FieldsMap{
			"error":       err.Error(),
			"sos_id":      assignReq.SOSID,
			"facility_id": assignReq.FacilityID,
			"request_id":  requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else if errorx.IsOfType(err, errorx.Validation) {
			httpStatus = http.StatusBadRequest
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	resp := sosEventToResponse(sosEvent)

	h.logger.Info(ctx, "Facility assigned to SOS event successfully", logger.FieldsMap{
		"sos_id":      sosEvent.ID.String(),
		"facility_id": facilityID.String(),
		"request_id":  requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// GetSOSByMother handles getting SOS events for a mother
func (h *SOSHandler) GetSOSByMother(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &GetSOSByMotherRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse get SOS by mother request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	getReq := req.(*GetSOSByMotherRequest)

	// Convert string ID to UUID
	motherID, err := uuid.Parse(getReq.MotherID)
	if err != nil {
		h.logger.Error(ctx, "Invalid mother ID", logger.FieldsMap{
			"error":      err.Error(),
			"mother_id":  getReq.MotherID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid mother ID", requestID)
		return
	}

	// Call service
	sosEvents, err := h.sosService.GetSOSEventsByMotherID(ctx, motherID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get SOS events for mother", logger.FieldsMap{
			"error":      err.Error(),
			"mother_id":  getReq.MotherID,
			"request_id": requestID,
		})

		var httpStatus int
		if errorx.IsOfType(err, errorx.NotFound) {
			httpStatus = http.StatusNotFound
		} else {
			httpStatus = http.StatusInternalServerError
		}

		response.WriteErrorResponse(w, httpStatus, err.Error(), requestID)
		return
	}

	// Prepare response
	sosResponses := make([]SOSResponse, 0, len(sosEvents))
	for _, sosEvent := range sosEvents {
		sosResponses = append(sosResponses, sosEventToResponse(sosEvent))
	}

	resp := SOSListResponse{
		SOSEvents: sosResponses,
	}

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// GetActiveSOSEvents handles getting all active SOS events
func (h *SOSHandler) GetActiveSOSEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Call service
	sosEvents, err := h.sosService.GetActiveSOSEvents(ctx)
	if err != nil {
		h.logger.Error(ctx, "Failed to get active SOS events", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})

		response.WriteErrorResponse(w, http.StatusInternalServerError, err.Error(), requestID)
		return
	}

	// Prepare response
	sosResponses := make([]SOSResponse, 0, len(sosEvents))
	for _, sosEvent := range sosEvents {
		sosResponses = append(sosResponses, sosEventToResponse(sosEvent))
	}

	resp := SOSListResponse{
		SOSEvents: sosResponses,
	}

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// Helper function to convert an SOS event to a response
func sosEventToResponse(sosEvent *model.SOSEvent) SOSResponse {
	var facilityIDStr *string
	if sosEvent.FacilityID != nil {
		idStr := sosEvent.FacilityID.String()
		facilityIDStr = &idStr
	}

	var ambulanceIDStr *string
	if sosEvent.AmbulanceID != nil {
		idStr := sosEvent.AmbulanceID.String()
		ambulanceIDStr = &idStr
	}

	var etaStr *string
	if sosEvent.ETA != nil {
		formattedETA := sosEvent.ETA.Format(time.RFC3339)
		etaStr = &formattedETA
	}

	return SOSResponse{
		ID:          sosEvent.ID.String(),
		MotherID:    sosEvent.MotherID.String(),
		ReportedBy:  sosEvent.ReportedBy.String(),
		Status:      string(sosEvent.Status),
		Nature:      string(sosEvent.Nature),
		Description: sosEvent.Description,
		Latitude:    sosEvent.Location.Latitude,
		Longitude:   sosEvent.Location.Longitude,
		CreatedAt:   sosEvent.CreatedAt,
		UpdatedAt:   sosEvent.UpdatedAt,
		FacilityID:  facilityIDStr,
		AmbulanceID: ambulanceIDStr,
		Priority:    sosEvent.Priority,
		ETA:         etaStr,
	}
}
