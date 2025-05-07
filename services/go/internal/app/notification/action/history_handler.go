package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/notification/history"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// GetNotificationHistoryRequest is the request for getting notification history
type GetNotificationHistoryRequest struct {
	UserID    string     `json:"user_id" validate:"required,uuid"`
	Type      string     `json:"type,omitempty"`
	Category  string     `json:"category,omitempty"`
	Channel   string     `json:"channel,omitempty"`
	Status    string     `json:"status,omitempty"`
	StartDate string     `json:"start_date,omitempty" validate:"omitempty,rfc3339"`
	EndDate   string     `json:"end_date,omitempty" validate:"omitempty,rfc3339"`
	Limit     int        `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Offset    int        `json:"offset,omitempty" validate:"omitempty,min=0"`
}

// MarkAsReadRequest is the request for marking a notification as read
type MarkAsReadRequest struct {
	NotificationID string `json:"notification_id" validate:"required,uuid"`
}

// RecordResponseRequest is the request for recording a response to a notification
type RecordResponseRequest struct {
	NotificationID string `json:"notification_id" validate:"required,uuid"`
	Response       string `json:"response" validate:"required"`
}

// CountUnreadRequest is the request for counting unread notifications
type CountUnreadRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// DeleteHistoryRequest is the request for deleting notification history
type DeleteHistoryRequest struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	OlderThan string `json:"older_than,omitempty" validate:"omitempty,rfc3339"`
}

// HistoryHandler handles notification history requests
type HistoryHandler struct {
	hasura.BaseActionHandler
	historyService *history.Service
	validator      *validation.Validator
	log            logger.Logger
}

// NewHistoryHandler creates a new notification history handler
func NewHistoryHandler(
	log logger.Logger,
	historyService *history.Service,
	validator *validation.Validator,
) *HistoryHandler {
	return &HistoryHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		historyService:    historyService,
		validator:         validator,
		log:               log,
	}
}

// GetNotificationHistory gets the notification history for a user
func (h *HistoryHandler) GetNotificationHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetNotificationHistoryRequest
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

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid user ID"))
		return
	}

	// Set up filter
	filter := &history.NotificationFilter{
		UserID:  &userID,
		Limit:   req.Limit,
		Offset:  req.Offset,
	}

	// Add optional filters
	if req.Type != "" {
		typeStr := req.Type
		filter.Type = &typeStr
	}
	if req.Category != "" {
		categoryStr := req.Category
		filter.Category = &categoryStr
	}
	if req.Channel != "" {
		channelStr := req.Channel
		filter.Channel = &channelStr
	}
	if req.Status != "" {
		statusStr := req.Status
		filter.Status = &statusStr
	}
	if req.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			h.log.Error("Invalid start date", logger.Fields{
				"request_id": reqID,
				"error":      err.Error(),
				"start_date": req.StartDate,
			})
			response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid start date format. Use RFC3339 format."))
			return
		}
		filter.StartDate = &startDate
	}
	if req.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			h.log.Error("Invalid end date", logger.Fields{
				"request_id": reqID,
				"error":      err.Error(),
				"end_date":   req.EndDate,
			})
			response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid end date format. Use RFC3339 format."))
			return
		}
		filter.EndDate = &endDate
	}

	// Get notification history
	notifications, err := h.historyService.GetNotificationHistory(ctx, filter)
	if err != nil {
		h.log.Error("Failed to get notification history", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got notification history", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
		"count":      len(notifications),
	})

	response.WriteJSONResponse(w, reqID, notifications)
}

// MarkAsRead marks a notification as read
func (h *HistoryHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req MarkAsReadRequest
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

	// Parse notification ID
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		h.log.Error("Invalid notification ID", logger.Fields{
			"request_id":      reqID,
			"error":           err.Error(),
			"notification_id": req.NotificationID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid notification ID"))
		return
	}

	// Mark as read
	if err := h.historyService.MarkNotificationAsRead(ctx, notificationID); err != nil {
		h.log.Error("Failed to mark notification as read", logger.Fields{
			"request_id":      reqID,
			"error":           err.Error(),
			"notification_id": req.NotificationID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Marked notification as read", logger.Fields{
		"request_id":      reqID,
		"notification_id": req.NotificationID,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

// RecordResponse records a response to a notification
func (h *HistoryHandler) RecordResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req RecordResponseRequest
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

	// Parse notification ID
	notificationID, err := uuid.Parse(req.NotificationID)
	if err != nil {
		h.log.Error("Invalid notification ID", logger.Fields{
			"request_id":      reqID,
			"error":           err.Error(),
			"notification_id": req.NotificationID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid notification ID"))
		return
	}

	// Record response
	if err := h.historyService.RecordNotificationResponse(ctx, notificationID, req.Response); err != nil {
		h.log.Error("Failed to record notification response", logger.Fields{
			"request_id":      reqID,
			"error":           err.Error(),
			"notification_id": req.NotificationID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Recorded notification response", logger.Fields{
		"request_id":      reqID,
		"notification_id": req.NotificationID,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

// CountUnread counts the number of unread notifications for a user
func (h *HistoryHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CountUnreadRequest
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

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid user ID"))
		return
	}

	// Count unread
	count, err := h.historyService.CountUnreadNotifications(ctx, userID)
	if err != nil {
		h.log.Error("Failed to count unread notifications", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Counted unread notifications", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
		"count":      count,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Count int `json:"count"`
	}{
		Count: count,
	})
}

// DeleteHistory deletes notification history for a user
func (h *HistoryHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req DeleteHistoryRequest
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

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.log.Error("Invalid user ID", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid user ID"))
		return
	}

	// Parse older than if provided
	var olderThan *time.Time
	if req.OlderThan != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.OlderThan)
		if err != nil {
			h.log.Error("Invalid older than date", logger.Fields{
				"request_id": reqID,
				"error":      err.Error(),
				"older_than": req.OlderThan,
			})
			response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid older than format. Use RFC3339 format."))
			return
		}
		olderThan = &parsedTime
	}

	// Delete history
	if err := h.historyService.DeleteNotificationHistory(ctx, userID, olderThan); err != nil {
		h.log.Error("Failed to delete notification history", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Deleted notification history", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
		"older_than": req.OlderThan,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}
