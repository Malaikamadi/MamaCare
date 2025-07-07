package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/notification/sms"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// SMSNotificationRequest is the request for sending an SMS notification
type SMSNotificationRequest struct {
	UserID     string      `json:"user_id" validate:"required,uuid"`
	Body       string      `json:"body" validate:"required"`
	Data       interface{} `json:"data,omitempty"`
	SenderID   string      `json:"sender_id,omitempty"`
	Priority   string      `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider   string      `json:"provider,omitempty" validate:"omitempty,oneof=twilio africas_talking"`
	MediaURL   string      `json:"media_url,omitempty"`
}

// MultiSMSNotificationRequest is the request for sending an SMS to multiple users
type MultiSMSNotificationRequest struct {
	UserIDs    []string    `json:"user_ids" validate:"required,min=1,dive,uuid"`
	Body       string      `json:"body" validate:"required"`
	Data       interface{} `json:"data,omitempty"`
	SenderID   string      `json:"sender_id,omitempty"`
	Priority   string      `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider   string      `json:"provider,omitempty" validate:"omitempty,oneof=twilio africas_talking"`
	MediaURL   string      `json:"media_url,omitempty"`
}

// DirectSMSRequest is the request for sending an SMS directly to a phone number
type DirectSMSRequest struct {
	PhoneNumber string      `json:"phone_number" validate:"required,e164"`
	Body        string      `json:"body" validate:"required"`
	Data        interface{} `json:"data,omitempty"`
	SenderID    string      `json:"sender_id,omitempty"`
	Priority    string      `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider    string      `json:"provider,omitempty" validate:"omitempty,oneof=twilio africas_talking"`
	MediaURL    string      `json:"media_url,omitempty"`
}

// ScheduleSMSRequest is the request for scheduling an SMS
type ScheduleSMSRequest struct {
	UserID      string      `json:"user_id" validate:"required,uuid"`
	Body        string      `json:"body" validate:"required"`
	Data        interface{} `json:"data,omitempty"`
	SenderID    string      `json:"sender_id,omitempty"`
	Priority    string      `json:"priority,omitempty" validate:"omitempty,oneof=normal high critical"`
	Provider    string      `json:"provider,omitempty" validate:"omitempty,oneof=twilio africas_talking"`
	MediaURL    string      `json:"media_url,omitempty"`
	ScheduleAt  string      `json:"schedule_at" validate:"required,rfc3339"`
}

// SMSHandler handles SMS notification requests
type SMSHandler struct {
	hasura.BaseActionHandler
	smsService *sms.Service
	validator  *validation.Validator
	log        logger.Logger
}

// NewSMSHandler creates a new SMS notification handler
func NewSMSHandler(
	log logger.Logger,
	smsService *sms.Service,
	validator *validation.Validator,
) *SMSHandler {
	return &SMSHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		smsService:        smsService,
		validator:         validator,
		log:               log,
	}
}

// SendSMS sends an SMS notification to a user
func (h *SMSHandler) SendSMS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req SMSNotificationRequest
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

	// Create message payload
	payload := &sms.MessagePayload{
		Body:      req.Body,
		Data:      req.Data,
		SenderID:  req.SenderID,
		Priority:  sms.SMSPriority(req.Priority),
		MediaURL:  req.MediaURL,
	}

	// Parse provider
	var provider sms.SMSProvider
	if req.Provider != "" {
		provider = sms.SMSProvider(req.Provider)
	}

	// Send SMS
	result, err := h.smsService.SendToUser(ctx, userID, payload, provider)
	if err != nil {
		h.log.Error("Failed to send SMS", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Sent SMS", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
		"status":     result.Status,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// SendMultiSMS sends an SMS notification to multiple users
func (h *SMSHandler) SendMultiSMS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req MultiSMSNotificationRequest
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

	// Create message payload
	payload := &sms.MessagePayload{
		Body:      req.Body,
		Data:      req.Data,
		SenderID:  req.SenderID,
		Priority:  sms.SMSPriority(req.Priority),
		MediaURL:  req.MediaURL,
	}

	// Parse provider
	var provider sms.SMSProvider
	if req.Provider != "" {
		provider = sms.SMSProvider(req.Provider)
	}

	// Send SMS
	results, err := h.smsService.SendToMultipleUsers(ctx, userIDs, payload, provider)
	if err != nil {
		h.log.Error("Failed to send multiple SMS", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_count": len(req.UserIDs),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Convert results to string-keyed map for JSON
	stringResults := make(map[string]*sms.SMSResult, len(results))
	for id, result := range results {
		stringResults[id.String()] = result
	}

	h.log.Info("Sent multiple SMS", logger.Fields{
		"request_id": reqID,
		"user_count": len(req.UserIDs),
	})

	response.WriteJSONResponse(w, reqID, stringResults)
}

// SendDirectSMS sends an SMS directly to a phone number
func (h *SMSHandler) SendDirectSMS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req DirectSMSRequest
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

	// Create message payload
	payload := &sms.MessagePayload{
		Body:      req.Body,
		Data:      req.Data,
		SenderID:  req.SenderID,
		Priority:  sms.SMSPriority(req.Priority),
		MediaURL:  req.MediaURL,
	}

	// Parse provider
	var provider sms.SMSProvider
	if req.Provider != "" {
		provider = sms.SMSProvider(req.Provider)
	}

	// Send SMS
	result, err := h.smsService.SendToNumber(ctx, req.PhoneNumber, payload, provider)
	if err != nil {
		h.log.Error("Failed to send direct SMS", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"phone":      req.PhoneNumber,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Sent direct SMS", logger.Fields{
		"request_id": reqID,
		"phone":      req.PhoneNumber,
		"status":     result.Status,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// ScheduleSMS schedules an SMS for future delivery
func (h *SMSHandler) ScheduleSMS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req ScheduleSMSRequest
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

	// Create message payload
	payload := &sms.MessagePayload{
		Body:        req.Body,
		Data:        req.Data,
		SenderID:    req.SenderID,
		Priority:    sms.SMSPriority(req.Priority),
		MediaURL:    req.MediaURL,
		ScheduleFor: &scheduleAt,
	}

	// Parse provider
	var provider sms.SMSProvider
	if req.Provider != "" {
		provider = sms.SMSProvider(req.Provider)
	}

	// Schedule SMS
	result, err := h.smsService.SendToUser(ctx, userID, payload, provider)
	if err != nil {
		h.log.Error("Failed to schedule SMS", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Scheduled SMS", logger.Fields{
		"request_id":  reqID,
		"user_id":     req.UserID,
		"schedule_at": scheduleAt.Format(time.RFC3339),
		"status":      result.Status,
	})

	response.WriteJSONResponse(w, reqID, result)
}

// ValidateE164 validates a phone number in E.164 format
func (h *SMSHandler) ValidateE164(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req struct {
		PhoneNumber string `json:"phone_number" validate:"required"`
	}
	if err := h.ParseRequest(r, &req); err != nil {
		h.log.Error("Failed to parse request", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Validate phone number
	isValid := h.smsService.ValidateE164Format(req.PhoneNumber)

	result := struct {
		Valid bool `json:"valid"`
	}{
		Valid: isValid,
	}

	h.log.Info("Validated phone number", logger.Fields{
		"request_id": reqID,
		"valid":      isValid,
	})

	response.WriteJSONResponse(w, reqID, result)
}
