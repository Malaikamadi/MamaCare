package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/notification/push"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// PushNotificationRequest is the request for sending a push notification
type PushNotificationRequest struct {
	UserID     string                  `json:"user_id" validate:"required,uuid"`
	Title      string                  `json:"title" validate:"required"`
	Body       string                  `json:"body" validate:"required"`
	Data       map[string]interface{}  `json:"data,omitempty"`
	Badge      int                     `json:"badge,omitempty"`
	Sound      string                  `json:"sound,omitempty"`
	ChannelID  string                  `json:"channel_id,omitempty"`
	Priority   string                  `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider   string                  `json:"provider,omitempty" validate:"omitempty,oneof=expo firebase"`
}

// MultiPushNotificationRequest is the request for sending a push notification to multiple users
type MultiPushNotificationRequest struct {
	UserIDs    []string               `json:"user_ids" validate:"required,min=1,dive,uuid"`
	Title      string                 `json:"title" validate:"required"`
	Body       string                 `json:"body" validate:"required"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Badge      int                    `json:"badge,omitempty"`
	Sound      string                 `json:"sound,omitempty"`
	ChannelID  string                 `json:"channel_id,omitempty"`
	Priority   string                 `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider   string                 `json:"provider,omitempty" validate:"omitempty,oneof=expo firebase"`
}

// TopicPushNotificationRequest is the request for sending a push notification to a topic
type TopicPushNotificationRequest struct {
	Topic      string                 `json:"topic" validate:"required"`
	Title      string                 `json:"title" validate:"required"`
	Body       string                 `json:"body" validate:"required"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Badge      int                    `json:"badge,omitempty"`
	Sound      string                 `json:"sound,omitempty"`
	ChannelID  string                 `json:"channel_id,omitempty"`
	Priority   string                 `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider   string                 `json:"provider,omitempty" validate:"omitempty,oneof=expo firebase"`
}

// ValidateTokenRequest is the request for validating a push token
type ValidateTokenRequest struct {
	Token    string `json:"token" validate:"required"`
	Provider string `json:"provider" validate:"required,oneof=expo firebase"`
}

// SchedulePushRequest is the request for scheduling a push notification
type SchedulePushRequest struct {
	UserID     string                 `json:"user_id" validate:"required,uuid"`
	Title      string                 `json:"title" validate:"required"`
	Body       string                 `json:"body" validate:"required"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Badge      int                    `json:"badge,omitempty"`
	Sound      string                 `json:"sound,omitempty"`
	ChannelID  string                 `json:"channel_id,omitempty"`
	Priority   string                 `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider   string                 `json:"provider,omitempty" validate:"omitempty,oneof=expo firebase"`
	ScheduleAt string                 `json:"schedule_at" validate:"required,rfc3339"`
}

// PushHandler handles push notification requests
type PushHandler struct {
	hasura.BaseActionHandler
	pushService     *push.Service
	validator       *validation.Validator
	log             logger.Logger
}

// NewPushHandler creates a new push notification handler
func NewPushHandler(
	log logger.Logger,
	pushService *push.Service,
	validator *validation.Validator,
) *PushHandler {
	return &PushHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		pushService:       pushService,
		validator:         validator,
		log:               log,
	}
}

// SendPushNotification sends a push notification to a user
func (h *PushHandler) SendPushNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req PushNotificationRequest
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

	// Create notification payload
	payload := &push.NotificationPayload{
		Title:      req.Title,
		Body:       req.Body,
		Data:       req.Data,
		Badge:      req.Badge,
		Sound:      req.Sound,
		ChannelID:  req.ChannelID,
		Priority:   push.NotificationPriority(req.Priority),
	}

	// Parse provider
	var provider push.PushProvider
	if req.Provider != "" {
		provider = push.PushProvider(req.Provider)
	}

	// Send notification
	result, err := h.pushService.SendToUser(ctx, userID, payload, provider)
	if err != nil {
		h.log.Error("Failed to send push notification", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Sent push notification", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
		"status":     result.Status,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// SendMultiPushNotification sends a push notification to multiple users
func (h *PushHandler) SendMultiPushNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req MultiPushNotificationRequest
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

	// Parse user IDs
	userIDs := make([]uuid.UUID, len(req.UserIDs))
	for i, id := range req.UserIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			h.log.Error("Invalid user ID", logger.Fields{
				"request_id": reqID,
				"error":      err.Error(),
				"user_id":    id,
			})
			response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid user ID: "+id))
			return
		}
		userIDs[i] = parsed
	}

	// Create notification payload
	payload := &push.NotificationPayload{
		Title:      req.Title,
		Body:       req.Body,
		Data:       req.Data,
		Badge:      req.Badge,
		Sound:      req.Sound,
		ChannelID:  req.ChannelID,
		Priority:   push.NotificationPriority(req.Priority),
	}

	// Parse provider
	var provider push.PushProvider
	if req.Provider != "" {
		provider = push.PushProvider(req.Provider)
	}

	// Send notifications
	results, err := h.pushService.SendToMultipleUsers(ctx, userIDs, payload, provider)
	if err != nil {
		h.log.Error("Failed to send push notifications", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_count": len(req.UserIDs),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Convert results to string-keyed map for JSON
	stringResults := make(map[string]*push.PushResult, len(results))
	for id, result := range results {
		stringResults[id.String()] = result
	}

	h.log.Info("Sent push notifications", logger.Fields{
		"request_id": reqID,
		"user_count": len(req.UserIDs),
	})

	response.WriteJSONResponse(w, reqID, stringResults)
}

// SendTopicPushNotification sends a push notification to a topic
func (h *PushHandler) SendTopicPushNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req TopicPushNotificationRequest
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

	// Create notification payload
	payload := &push.NotificationPayload{
		Title:      req.Title,
		Body:       req.Body,
		Data:       req.Data,
		Badge:      req.Badge,
		Sound:      req.Sound,
		ChannelID:  req.ChannelID,
		Priority:   push.NotificationPriority(req.Priority),
	}

	// Parse provider
	var provider push.PushProvider
	if req.Provider != "" {
		provider = push.PushProvider(req.Provider)
	}

	// Send notification
	result, err := h.pushService.SendToTopic(ctx, req.Topic, payload, provider)
	if err != nil {
		h.log.Error("Failed to send topic push notification", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"topic":      req.Topic,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Sent topic push notification", logger.Fields{
		"request_id": reqID,
		"topic":      req.Topic,
		"status":     result.Status,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// ValidatePushToken validates a push token
func (h *PushHandler) ValidatePushToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req ValidateTokenRequest
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

	var isValid bool
	if req.Provider == "expo" {
		isValid = h.pushService.ValidateExpoToken(req.Token)
	} else {
		isValid = h.pushService.ValidateFirebaseToken(req.Token)
	}

	result := struct {
		Valid bool `json:"valid"`
	}{
		Valid: isValid,
	}

	h.log.Info("Validated push token", logger.Fields{
		"request_id": reqID,
		"provider":   req.Provider,
		"valid":      isValid,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// SchedulePushNotification schedules a push notification for future delivery
func (h *PushHandler) SchedulePushNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req SchedulePushRequest
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

	// Parse schedule time
	scheduleAt, err := time.Parse(time.RFC3339, req.ScheduleAt)
	if err != nil {
		h.log.Error("Invalid schedule time", logger.Fields{
			"request_id":  reqID,
			"error":       err.Error(),
			"schedule_at": req.ScheduleAt,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid schedule time format. Use RFC3339 format."))
		return
	}

	// Create notification payload
	payload := &push.NotificationPayload{
		Title:      req.Title,
		Body:       req.Body,
		Data:       req.Data,
		Badge:      req.Badge,
		Sound:      req.Sound,
		ChannelID:  req.ChannelID,
		Priority:   push.NotificationPriority(req.Priority),
	}

	// Parse provider
	var provider push.PushProvider
	if req.Provider != "" {
		provider = push.PushProvider(req.Provider)
	}

	// Schedule notification
	result := struct {
		ScheduledAt time.Time `json:"scheduled_at"`
		Status      string    `json:"status"`
	}{
		ScheduledAt: scheduleAt,
		Status:      "scheduled",
	}

	h.log.Info("Scheduled push notification", logger.Fields{
		"request_id":  reqID,
		"user_id":     req.UserID,
		"schedule_at": scheduleAt.Format(time.RFC3339),
	})

	response.WriteJSONResponse(w, reqID, result)
}
