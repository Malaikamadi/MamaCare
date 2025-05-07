package action

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/incognito25/mamacare/services/go/internal/app/emergency/tracking"
	"github.com/incognito25/mamacare/services/go/internal/domain/model"
	"github.com/incognito25/mamacare/services/go/internal/errorx"
	"github.com/incognito25/mamacare/services/go/internal/logger"
	"github.com/incognito25/mamacare/services/go/internal/port/hasura"
	"github.com/incognito25/mamacare/services/go/internal/port/response"
)

// TrackingHandler handles all emergency tracking related actions
type TrackingHandler struct {
	hasura.BaseActionHandler
	trackingService *tracking.Service
	logger          logger.Logger
}

// GetEmergencyStatusRequest defines the request payload for getting emergency status
type GetEmergencyStatusRequest struct {
	SOSID string `json:"sos_id"`
}

// GetAmbulanceLocationRequest defines the request payload for getting ambulance location
type GetAmbulanceLocationRequest struct {
	AmbulanceID string `json:"ambulance_id"`
}

// RecordAmbulanceArrivalRequest defines the request payload for recording ambulance arrival
type RecordAmbulanceArrivalRequest struct {
	SOSID string `json:"sos_id"`
}

// TrackRouteRequest defines the request payload for tracking an emergency route
type TrackRouteRequest struct {
	SOSID string `json:"sos_id"`
}

// UpdateETARequest defines the request payload for updating ETA based on traffic
type UpdateETARequest struct {
	SOSID               string `json:"sos_id"`
	TrafficDelayMinutes int    `json:"traffic_delay_minutes"`
}

// RecordStatusUpdateRequest defines the request payload for recording a status update
type RecordStatusUpdateRequest struct {
	SOSID       string `json:"sos_id"`
	UpdateType  string `json:"update_type"`
	Description string `json:"description"`
}

// GetETARequest defines the request payload for getting estimated time of arrival
type GetETARequest struct {
	SOSID string `json:"sos_id"`
}

// RouteResponse defines the response for a tracking route
type RouteResponse struct {
	Distance     float64       `json:"distance_km"`
	Duration     int           `json:"duration_minutes"`
	StartPoint   LocationPoint `json:"start_point"`
	EndPoint     LocationPoint `json:"end_point"`
	EncodedPath  string        `json:"encoded_path,omitempty"`
	Instructions []string      `json:"instructions,omitempty"`
}

// LocationPoint represents a point on a route
type LocationPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
}

// ETAResponse defines the response for an ETA request
type ETAResponse struct {
	ETA         string `json:"eta"`
	ETAMinutes  int    `json:"eta_minutes"`
	AmbulanceID string `json:"ambulance_id,omitempty"`
}

// StatusUpdateResponse defines the response for a status update
type StatusUpdateResponse struct {
	ID          string     `json:"id"`
	SOSEventID  string     `json:"sos_event_id"`
	UpdateType  string     `json:"update_type"`
	Description string     `json:"description"`
	Timestamp   time.Time  `json:"timestamp"`
	Location    *Location  `json:"location,omitempty"`
	ETA         *time.Time `json:"eta,omitempty"`
}

// LocationResponse defines the response for a location request
type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// NewTrackingHandler creates a new tracking handler
func NewTrackingHandler(trackingService *tracking.Service, logger logger.Logger) *TrackingHandler {
	return &TrackingHandler{
		trackingService: trackingService,
		logger:          logger,
	}
}

// GetEmergencyStatus handles getting the status of an emergency
func (h *TrackingHandler) GetEmergencyStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &GetEmergencyStatusRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse get emergency status request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	getReq := req.(*GetEmergencyStatusRequest)

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
	sosEvent, err := h.trackingService.GetEmergencyStatus(ctx, sosID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get emergency status", logger.FieldsMap{
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

// GetAmbulanceLocation handles getting the location of an ambulance
func (h *TrackingHandler) GetAmbulanceLocation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &GetAmbulanceLocationRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse get ambulance location request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	getReq := req.(*GetAmbulanceLocationRequest)

	// Convert string ID to UUID
	ambulanceID, err := uuid.Parse(getReq.AmbulanceID)
	if err != nil {
		h.logger.Error(ctx, "Invalid ambulance ID", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": getReq.AmbulanceID,
			"request_id":   requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid ambulance ID", requestID)
		return
	}

	// Call service
	location, err := h.trackingService.GetAmbulanceLocation(ctx, ambulanceID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ambulance location", logger.FieldsMap{
			"error":        err.Error(),
			"ambulance_id": getReq.AmbulanceID,
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
	resp := LocationResponse{
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
	}

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// TrackRoute handles tracking the route for an emergency
func (h *TrackingHandler) TrackRoute(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &TrackRouteRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse track route request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	trackReq := req.(*TrackRouteRequest)

	// Convert string ID to UUID
	sosID, err := uuid.Parse(trackReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     trackReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	// Call service
	route, err := h.trackingService.TrackRoute(ctx, sosID)
	if err != nil {
		h.logger.Error(ctx, "Failed to track route", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     trackReq.SOSID,
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
	resp := RouteResponse{
		Distance: route.DistanceKm,
		Duration: int(route.DurationMinutes),
		StartPoint: LocationPoint{
			Latitude:  route.StartPoint.Latitude,
			Longitude: route.StartPoint.Longitude,
			Name:      route.StartPoint.Name,
		},
		EndPoint: LocationPoint{
			Latitude:  route.EndPoint.Latitude,
			Longitude: route.EndPoint.Longitude,
			Name:      route.EndPoint.Name,
		},
		EncodedPath:  route.EncodedPath,
		Instructions: route.Instructions,
	}

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// UpdateETABasedOnTraffic handles updating the ETA based on traffic
func (h *TrackingHandler) UpdateETABasedOnTraffic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &UpdateETARequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse update ETA request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	updateReq := req.(*UpdateETARequest)

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

	// Call service
	newETA, err := h.trackingService.UpdateETABasedOnTraffic(ctx, sosID, updateReq.TrafficDelayMinutes)
	if err != nil {
		h.logger.Error(ctx, "Failed to update ETA based on traffic", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     updateReq.SOSID,
			"delay":      updateReq.TrafficDelayMinutes,
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
	resp := ETAResponse{
		ETA:        newETA.Format(time.RFC3339),
		ETAMinutes: int(time.Until(*newETA).Minutes()),
	}

	h.logger.Info(ctx, "Updated ETA based on traffic", logger.FieldsMap{
		"sos_id":      updateReq.SOSID,
		"delay":       updateReq.TrafficDelayMinutes,
		"new_eta":     newETA.Format(time.RFC3339),
		"eta_minutes": resp.ETAMinutes,
		"request_id":  requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// RecordAmbulanceArrival handles recording that an ambulance has arrived
func (h *TrackingHandler) RecordAmbulanceArrival(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &RecordAmbulanceArrivalRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse record arrival request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	arrivalReq := req.(*RecordAmbulanceArrivalRequest)

	// Convert string ID to UUID
	sosID, err := uuid.Parse(arrivalReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     arrivalReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	// Call service
	if err := h.trackingService.RecordAmbulanceArrival(ctx, sosID); err != nil {
		h.logger.Error(ctx, "Failed to record ambulance arrival", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     arrivalReq.SOSID,
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

	h.logger.Info(ctx, "Recorded ambulance arrival", logger.FieldsMap{
		"sos_id":     arrivalReq.SOSID,
		"request_id": requestID,
	})

	// Return success without body
	response.WriteJSONResponse(w, http.StatusOK, map[string]bool{"success": true}, requestID)
}

// RecordEmergencyStatusUpdate handles recording a status update for an emergency
func (h *TrackingHandler) RecordEmergencyStatusUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &RecordStatusUpdateRequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse record status update request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	updateReq := req.(*RecordStatusUpdateRequest)

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

	// Call service
	trackingUpdate, err := h.trackingService.RecordEmergencyStatusUpdate(ctx, sosID, updateReq.UpdateType, updateReq.Description)
	if err != nil {
		h.logger.Error(ctx, "Failed to record emergency status update", logger.FieldsMap{
			"error":       err.Error(),
			"sos_id":      updateReq.SOSID,
			"update_type": updateReq.UpdateType,
			"request_id":  requestID,
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
	var locationResp *Location
	if trackingUpdate.Location != nil {
		locationResp = &Location{
			Latitude:  trackingUpdate.Location.Latitude,
			Longitude: trackingUpdate.Location.Longitude,
		}
	}

	resp := StatusUpdateResponse{
		ID:          trackingUpdate.ID.String(),
		SOSEventID:  trackingUpdate.SOSEventID.String(),
		UpdateType:  trackingUpdate.UpdateType,
		Description: trackingUpdate.Description,
		Timestamp:   trackingUpdate.Timestamp,
		Location:    locationResp,
		ETA:         trackingUpdate.ETA,
	}

	h.logger.Info(ctx, "Recorded emergency status update", logger.FieldsMap{
		"update_id":   trackingUpdate.ID.String(),
		"sos_id":      updateReq.SOSID,
		"update_type": updateReq.UpdateType,
		"request_id":  requestID,
	})

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}

// GetEstimatedTimeOfArrival handles getting the ETA for an emergency
func (h *TrackingHandler) GetEstimatedTimeOfArrival(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := response.GetRequestID(ctx)

	// Parse request
	req, err := h.ParseRequest(r, &GetETARequest{})
	if err != nil {
		h.logger.Error(ctx, "Failed to parse get ETA request", logger.FieldsMap{
			"error":      err.Error(),
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload", requestID)
		return
	}

	etaReq := req.(*GetETARequest)

	// Convert string ID to UUID
	sosID, err := uuid.Parse(etaReq.SOSID)
	if err != nil {
		h.logger.Error(ctx, "Invalid SOS ID", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     etaReq.SOSID,
			"request_id": requestID,
		})
		response.WriteErrorResponse(w, http.StatusBadRequest, "Invalid SOS ID", requestID)
		return
	}

	// Call services to get ETA and minutes
	eta, err := h.trackingService.GetEstimatedTimeOfArrival(ctx, sosID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ETA", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     etaReq.SOSID,
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

	etaMinutes, err := h.trackingService.GetETAMinutes(ctx, sosID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get ETA minutes", logger.FieldsMap{
			"error":      err.Error(),
			"sos_id":     etaReq.SOSID,
			"request_id": requestID,
		})
		// Continue despite error, just use 0 as minutes
		etaMinutes = 0
	}

	// Get SOS event to extract ambulance ID
	sosEvent, err := h.trackingService.GetEmergencyStatus(ctx, sosID)
	var ambulanceIDStr string
	if err == nil && sosEvent.AmbulanceID != nil {
		ambulanceIDStr = sosEvent.AmbulanceID.String()
	}

	// Prepare response
	resp := ETAResponse{
		ETA:         eta.Format(time.RFC3339),
		ETAMinutes:  etaMinutes,
		AmbulanceID: ambulanceIDStr,
	}

	response.WriteJSONResponse(w, http.StatusOK, resp, requestID)
}
