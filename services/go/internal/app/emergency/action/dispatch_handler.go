package action

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/app/emergency/dispatch"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
	"github.com/incognito25/mamacare/services/go/internal/errorx"
	"github.com/incognito25/mamacare/services/go/internal/logger"
	"github.com/incognito25/mamacare/services/go/internal/port/hasura"
	"github.com/incognito25/mamacare/services/go/internal/port/response"
)

// DispatchHandler handles all ambulance dispatch related actions
type DispatchHandler struct {
	hasura.BaseActionHandler
	dispatchService *dispatch.Service
	logger          logger.Logger
}

// DispatchAmbulanceRequest defines the request payload for dispatching an ambulance
type DispatchAmbulanceRequest struct {
	SOSID       string `json:"sos_id"`
	AmbulanceID string `json:"ambulance_id"`
}

// FindSuitableAmbulancesRequest defines the request payload for finding suitable ambulances
type FindSuitableAmbulancesRequest struct {
	SOSID      string `json:"sos_id"`
	MaxResults int    `json:"max_results"`
}

// UpdateAmbulanceStatusRequest defines the request payload for updating ambulance status
type UpdateAmbulanceStatusRequest struct {
	AmbulanceID string `json:"ambulance_id"`
	Status      string `json:"status"`
}

// UpdateAmbulanceLocationRequest defines the request payload for updating ambulance location
type UpdateAmbulanceLocationRequest struct {
	AmbulanceID string  `json:"ambulance_id"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// AmbulanceResponse defines a standard ambulance response
type AmbulanceResponse struct {
	ID            string     `json:"id"`
	CallSign      string     `json:"call_sign"`
	VehicleID     string     `json:"vehicle_id"`
	AmbulanceType string     `json:"ambulance_type"`
	Capacity      int        `json:"capacity"`
	Status        string     `json:"status"`
	CurrentSOSID  *string    `json:"current_sos_id,omitempty"`
	FacilityID    string     `json:"facility_id"`
	Location      *Location  `json:"location,omitempty"`
	Crew          []string   `json:"crew"`
	LastUpdated   time.Time  `json:"last_updated"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Location represents a geographic location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// AmbulanceListResponse defines the response payload for lists of ambulances
type AmbulanceListResponse struct {
	Ambulances []AmbulanceResponse `json:"ambulances"`
}

// NewDispatchHandler creates a new dispatch handler
func NewDispatchHandler(dispatchService *dispatch.Service, logger logger.Logger) *DispatchHandler {
	return &DispatchHandler{
		dispatchService: dispatchService,
		logger:          logger,
	}
}

// DispatchAmbulance handles dispatching an ambulance to an SOS event
func (h *DispatchHandler) DispatchAmbulance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &DispatchAmbulanceRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse dispatch ambulance request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	dispatchReq := req.(*DispatchAmbulanceRequest)

	// Convert string IDs to UUID
	sosID, err := uuid.Parse(dispatchReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     dispatchReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	ambulanceID, err := uuid.Parse(dispatchReq.AmbulanceID)
	if err != nil {
		h.logger.Error(ctx, "Invalid ambulance ID", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": dispatchReq.AmbulanceID,
			"request_id":   requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid ambulance ID", requestID)
		return
	}

	// Call service
	sosEvent, err := h.dispatchService.DispatchAmbulance(ctx, sosID, ambulanceID)
	if err != nil {
		h.logger.Error(ctx, "Failed to dispatch ambulance", logger.FieldsMap{
			"error":        err.Error(),
			"sos_id":       dispatchReq.SOSID,
			"ambulance_id": dispatchReq.AmbulanceID,
			"request_id":   requestID,
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

	h.logger.Info(ctx, "Ambulance dispatched successfully", logger.FieldsMap{
		"sos_id":       sosEvent.ID.String(),
		"ambulance_id": ambulanceID.String(),
		"request_id":   requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// FindSuitableAmbulances handles finding suitable ambulances for an SOS event
func (h *DispatchHandler) FindSuitableAmbulances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &FindSuitableAmbulancesRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse find ambulances request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	findReq := req.(*FindSuitableAmbulancesRequest)

	// Convert string ID to UUID
	sosID, err := uuid.Parse(findReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     findReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	// Call service
	ambulances, err := h.dispatchService.FindSuitableAmbulances(ctx, sosID, findReq.MaxResults)
	if err != nil {
		h.logger.Error(ctx, "Failed to find suitable ambulances", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     findReq.SOSID,
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
	ambulanceResponses := make([]AmbulanceResponse, 0, len(ambulances))
	for _, ambulance := range ambulances {
		ambulanceResponses = append(ambulanceResponses, ambulanceToResponse(ambulance))
	}

	resp := AmbulanceListResponse{
		Ambulances: ambulanceResponses,
	}

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// UpdateAmbulanceStatus handles updating the status of an ambulance
func (h *DispatchHandler) UpdateAmbulanceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &UpdateAmbulanceStatusRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse update ambulance status request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	updateReq := req.(*UpdateAmbulanceStatusRequest)

	// Convert string ID to UUID
	ambulanceID, err := uuid.Parse(updateReq.AmbulanceID)
	if err != nil {
		h.logger.Error(ctx, "Invalid ambulance ID", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": updateReq.AmbulanceID,
			"request_id":   requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid ambulance ID", requestID)
		return
	}

	// Convert status string to enum
	var status model.AmbulanceStatus
	switch updateReq.Status {
	case "available":
		status = model.AmbulanceStatusAvailable
	case "dispatched":
		status = model.AmbulanceStatusDispatched
	case "en_route":
		status = model.AmbulanceStatusEnRoute
	case "arrived":
		status = model.AmbulanceStatusArrived
	case "returning":
		status = model.AmbulanceStatusReturning
	case "maintenance":
		status = model.AmbulanceStatusMaintenance
	default:
		h.logger.Error(ctx, "Invalid ambulance status", logger.FieldsMap{
			"status":     updateReq.Status,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid ambulance status", requestID)
		return
	}

	// Call service
	ambulance, err := h.dispatchService.UpdateAmbulanceStatus(ctx, ambulanceID, status)
	if err != nil {
		h.logger.Error(ctx, "Failed to update ambulance status", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": updateReq.AmbulanceID,
			"status":       updateReq.Status,
			"request_id":   requestID,
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
	resp := ambulanceToResponse(ambulance)

	h.logger.Info(ctx, "Ambulance status updated successfully", logger.FieldsMap{
		"ambulance_id": ambulance.ID.String(),
		"status":       string(ambulance.Status),
		"request_id":   requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// UpdateAmbulanceLocation handles updating the location of an ambulance
func (h *DispatchHandler) UpdateAmbulanceLocation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &UpdateAmbulanceLocationRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse update ambulance location request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	updateReq := req.(*UpdateAmbulanceLocationRequest)

	// Convert string ID to UUID
	ambulanceID, err := uuid.Parse(updateReq.AmbulanceID)
	if err != nil {
		h.logger.Error(ctx, "Invalid ambulance ID", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": updateReq.AmbulanceID,
			"request_id":   requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid ambulance ID", requestID)
		return
	}

	// Call service
	ambulance, err := h.dispatchService.UpdateAmbulanceLocation(ctx, ambulanceID, updateReq.Latitude, updateReq.Longitude)
	if err != nil {
		h.logger.Error(ctx, "Failed to update ambulance location", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": updateReq.AmbulanceID,
			"request_id":   requestID,
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
	resp := ambulanceToResponse(ambulance)

	h.logger.Info(ctx, "Ambulance location updated successfully", logger.FieldsMap{
		"ambulance_id": ambulance.ID.String(),
		"latitude":     updateReq.Latitude,
		"longitude":    updateReq.Longitude,
		"request_id":   requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// Helper function to convert an ambulance to a response
func ambulanceToResponse(ambulance *model.Ambulance) AmbulanceResponse {
	var currentSOSIDStr *string
	if ambulance.CurrentSOSID != nil {
		idStr := ambulance.CurrentSOSID.String()
		currentSOSIDStr = &idStr
	}

	var location *Location
	if ambulance.Location != nil {
		location = &Location{
			Latitude:  ambulance.Location.Latitude,
			Longitude: ambulance.Location.Longitude,
		}
	}

	crew := make([]string, 0, len(ambulance.Crew))
	for _, memberID := range ambulance.Crew {
		crew = append(crew, memberID.String())
	}

	return AmbulanceResponse{
		ID:            ambulance.ID.String(),
		CallSign:      ambulance.CallSign,
		VehicleID:     ambulance.VehicleID,
		AmbulanceType: string(ambulance.AmbulanceType),
		Capacity:      ambulance.Capacity,
		Status:        string(ambulance.Status),
		CurrentSOSID:  currentSOSIDStr,
		FacilityID:    ambulance.FacilityID.String(),
		Location:      location,
		Crew:          crew,
		LastUpdated:   ambulance.LastUpdated,
		CreatedAt:     ambulance.CreatedAt,
		UpdatedAt:     ambulance.UpdatedAt,
	}
}
