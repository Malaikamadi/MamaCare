package action

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/notification/preference"
	"github.com/mamacare/services/internal/port/hasura"
	"github.com/mamacare/services/internal/port/response"
	"github.com/mamacare/services/internal/port/validation"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// GetPreferencesRequest is the request for getting notification preferences
type GetPreferencesRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// UpdatePreferencesRequest is the request for updating notification preferences
type UpdatePreferencesRequest struct {
	UserID             string `json:"user_id" validate:"required,uuid"`
	SMSEnabled         *bool  `json:"sms_enabled,omitempty"`
	PushEnabled        *bool  `json:"push_enabled,omitempty"`
	EmailEnabled       *bool  `json:"email_enabled,omitempty"`
	QuietHoursEnabled  *bool  `json:"quiet_hours_enabled,omitempty"`
	QuietHoursStart    string `json:"quiet_hours_start,omitempty" validate:"omitempty,pattern=^([01]?[0-9]|2[0-3]):[0-5][0-9]$"`
	QuietHoursEnd      string `json:"quiet_hours_end,omitempty" validate:"omitempty,pattern=^([01]?[0-9]|2[0-3]):[0-5][0-9]$"`
	PreferredChannel   string `json:"preferred_channel,omitempty" validate:"omitempty,oneof=sms push email"`
	HealthAlertsEnabled *bool `json:"health_alerts_enabled,omitempty"`
	VisitReminders     *bool  `json:"visit_reminders,omitempty"`
	VisitReminderHours int    `json:"visit_reminder_hours,omitempty" validate:"omitempty,min=1,max=168"`
	VaccineReminders   *bool  `json:"vaccine_reminders,omitempty"`
}

// DisableAllRequest is the request for disabling all notifications
type DisableAllRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// CheckEnabledRequest is the request for checking if a notification type is enabled
type CheckEnabledRequest struct {
	UserID           string `json:"user_id" validate:"required,uuid"`
	NotificationType string `json:"notification_type" validate:"required,oneof=health_alert visit_reminder vaccine_reminder emergency chw_assignment general_info"`
}

// PreferenceHandler handles notification preference requests
type PreferenceHandler struct {
	hasura.BaseActionHandler
	preferenceService *preference.Service
	validator        *validation.Validator
	log              logger.Logger
}

// NewPreferenceHandler creates a new notification preference handler
func NewPreferenceHandler(
	log logger.Logger,
	preferenceService *preference.Service,
	validator *validation.Validator,
) *PreferenceHandler {
	return &PreferenceHandler{
		BaseActionHandler: hasura.BaseActionHandler{},
		preferenceService: preferenceService,
		validator:        validator,
		log:              log,
	}
}

// GetPreferences gets the notification preferences for a user
func (h *PreferenceHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetPreferencesRequest
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

	// Get preferences
	prefs, err := h.preferenceService.GetUserPreferences(ctx, userID)
	if err != nil {
		h.log.Error("Failed to get notification preferences", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got notification preferences", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
	})

	response.WriteJSONResponse(w, reqID, prefs)
}

// UpdatePreferences updates the notification preferences for a user
func (h *PreferenceHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req UpdatePreferencesRequest
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

	// Get current preferences
	currentPrefs, err := h.preferenceService.GetUserPreferences(ctx, userID)
	if err != nil {
		h.log.Error("Failed to get current notification preferences", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	// Update preferences with provided values
	updatePrefs := &preference.UserPreferences{
		ID:                 currentPrefs.ID,
		UserID:             currentPrefs.UserID,
		SMSEnabled:         currentPrefs.SMSEnabled,
		PushEnabled:        currentPrefs.PushEnabled,
		EmailEnabled:       currentPrefs.EmailEnabled,
		QuietHoursEnabled:  currentPrefs.QuietHoursEnabled,
		QuietHoursStart:    currentPrefs.QuietHoursStart,
		QuietHoursEnd:      currentPrefs.QuietHoursEnd,
		PreferredChannel:   currentPrefs.PreferredChannel,
		HealthAlertsEnabled: currentPrefs.HealthAlertsEnabled,
		VisitReminders:     currentPrefs.VisitReminders,
		VisitReminderHours: currentPrefs.VisitReminderHours,
		VaccineReminders:   currentPrefs.VaccineReminders,
		CreatedAt:          currentPrefs.CreatedAt,
		UpdatedAt:          time.Now(),
	}

	// Update only provided fields
	if req.SMSEnabled != nil {
		updatePrefs.SMSEnabled = *req.SMSEnabled
	}
	if req.PushEnabled != nil {
		updatePrefs.PushEnabled = *req.PushEnabled
	}
	if req.EmailEnabled != nil {
		updatePrefs.EmailEnabled = *req.EmailEnabled
	}
	if req.QuietHoursEnabled != nil {
		updatePrefs.QuietHoursEnabled = *req.QuietHoursEnabled
	}
	if req.QuietHoursStart != "" {
		updatePrefs.QuietHoursStart = req.QuietHoursStart
	}
	if req.QuietHoursEnd != "" {
		updatePrefs.QuietHoursEnd = req.QuietHoursEnd
	}
	if req.PreferredChannel != "" {
		updatePrefs.PreferredChannel = req.PreferredChannel
	}
	if req.HealthAlertsEnabled != nil {
		updatePrefs.HealthAlertsEnabled = *req.HealthAlertsEnabled
	}
	if req.VisitReminders != nil {
		updatePrefs.VisitReminders = *req.VisitReminders
	}
	if req.VisitReminderHours > 0 {
		updatePrefs.VisitReminderHours = req.VisitReminderHours
	}
	if req.VaccineReminders != nil {
		updatePrefs.VaccineReminders = *req.VaccineReminders
	}

	// Update preferences
	updatedPrefs, err := h.preferenceService.UpdateUserPreferences(ctx, userID, updatePrefs)
	if err != nil {
		h.log.Error("Failed to update notification preferences", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Updated notification preferences", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
	})

	response.WriteJSONResponse(w, reqID, updatedPrefs)
}

// DisableAllNotifications disables all notifications for a user
func (h *PreferenceHandler) DisableAllNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req DisableAllRequest
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

	// Disable all notifications
	if err := h.preferenceService.DisableAllNotifications(ctx, userID); err != nil {
		h.log.Error("Failed to disable all notifications", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Disabled all notifications", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

// CheckNotificationEnabled checks if a specific notification type is enabled for a user
func (h *PreferenceHandler) CheckNotificationEnabled(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req CheckEnabledRequest
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

	// Map notification type
	var notifType preference.NotificationType
	switch req.NotificationType {
	case "health_alert":
		notifType = preference.TypeHealthAlert
	case "visit_reminder":
		notifType = preference.TypeVisitReminder
	case "vaccine_reminder":
		notifType = preference.TypeVaccineReminder
	case "emergency":
		notifType = preference.TypeEmergency
	case "chw_assignment":
		notifType = preference.TypeCHWAssignment
	case "general_info":
		notifType = preference.TypeGeneralInfo
	default:
		h.log.Error("Invalid notification type", logger.Fields{
			"request_id":        reqID,
			"notification_type": req.NotificationType,
		})
		response.WriteErrorResponse(w, reqID, errorx.New(errorx.BadRequest, "Invalid notification type"))
		return
	}

	// Check if enabled
	enabled, err := h.preferenceService.IsNotificationEnabled(ctx, userID, notifType)
	if err != nil {
		h.log.Error("Failed to check if notification is enabled", logger.Fields{
			"request_id":        reqID,
			"error":             err.Error(),
			"user_id":           req.UserID,
			"notification_type": req.NotificationType,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Checked if notification is enabled", logger.Fields{
		"request_id":        reqID,
		"user_id":           req.UserID,
		"notification_type": req.NotificationType,
		"enabled":           enabled,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Enabled bool `json:"enabled"`
	}{
		Enabled: enabled,
	})
}

// GetPreferredChannel gets the preferred notification channel for a user
func (h *PreferenceHandler) GetPreferredChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := response.GetRequestID(ctx)

	var req GetPreferencesRequest
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

	// Get preferred channel
	channel, err := h.preferenceService.GetPreferredChannel(ctx, userID)
	if err != nil {
		h.log.Error("Failed to get preferred channel", logger.Fields{
			"request_id": reqID,
			"error":      err.Error(),
			"user_id":    req.UserID,
		})
		response.WriteErrorResponse(w, reqID, err)
		return
	}

	h.log.Info("Got preferred channel", logger.Fields{
		"request_id": reqID,
		"user_id":    req.UserID,
		"channel":    channel,
	})

	response.WriteJSONResponse(w, reqID, struct {
		Channel string `json:"channel"`
	}{
		Channel: channel,
	})
}
