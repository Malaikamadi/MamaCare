package preference

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// UserPreferences defines all notification preferences for a user
type UserPreferences struct {
	ID                  uuid.UUID `json:"id"`
	UserID              uuid.UUID `json:"user_id"`
	SMSEnabled          bool      `json:"sms_enabled"`
	PushEnabled         bool      `json:"push_enabled"`
	EmailEnabled        bool      `json:"email_enabled"`
	QuietHoursEnabled   bool      `json:"quiet_hours_enabled"`
	QuietHoursStart     string    `json:"quiet_hours_start,omitempty"` // Format: HH:MM in 24-hour
	QuietHoursEnd       string    `json:"quiet_hours_end,omitempty"`   // Format: HH:MM in 24-hour
	PreferredChannel    string    `json:"preferred_channel"`           // "sms", "push", "email"
	HealthAlertsEnabled bool      `json:"health_alerts_enabled"`
	VisitReminders      bool      `json:"visit_reminders"`
	VisitReminderHours  int       `json:"visit_reminder_hours"`
	VaccineReminders    bool      `json:"vaccine_reminders"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// NotificationType represents a specific type of notification
type NotificationType string

const (
	// TypeHealthAlert is for health-related alerts
	TypeHealthAlert NotificationType = "health_alert"
	// TypeVisitReminder is for upcoming visit reminders
	TypeVisitReminder NotificationType = "visit_reminder"
	// TypeVaccineReminder is for vaccine reminders
	TypeVaccineReminder NotificationType = "vaccine_reminder"
	// TypeEmergency is for emergency alerts
	TypeEmergency NotificationType = "emergency"
	// TypeCHWAssignment is for CHW assignment notifications
	TypeCHWAssignment NotificationType = "chw_assignment"
	// TypeGeneralInfo is for general information notifications
	TypeGeneralInfo NotificationType = "general_info"
)

// Service handles user notification preferences
type Service struct {
	notifPrefsRepo repository.NotificationPreferenceRepository
	userRepo       repository.UserRepository
	log            logger.Logger
}

// NewService creates a new notification preferences service
func NewService(
	notifPrefsRepo repository.NotificationPreferenceRepository,
	userRepo repository.UserRepository,
	log logger.Logger,
) *Service {
	return &Service{
		notifPrefsRepo: notifPrefsRepo,
		userRepo:       userRepo,
		log:            log,
	}
}

// GetUserPreferences retrieves notification preferences for a user
func (s *Service) GetUserPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*UserPreferences, error) {
	prefs, err := s.notifPrefsRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If preferences don't exist, create default preferences
		if errorx.IsNotFound(err) {
			return s.CreateDefaultPreferences(ctx, userID)
		}
		return nil, errorx.Wrap(err, "failed to get notification preferences")
	}

	return &UserPreferences{
		ID:                  prefs.ID,
		UserID:              prefs.UserID,
		SMSEnabled:          prefs.SMSEnabled,
		PushEnabled:         prefs.PushEnabled,
		EmailEnabled:        prefs.EmailEnabled,
		QuietHoursEnabled:   prefs.QuietHoursEnabled,
		QuietHoursStart:     prefs.QuietHoursStart,
		QuietHoursEnd:       prefs.QuietHoursEnd,
		PreferredChannel:    prefs.PreferredChannel,
		HealthAlertsEnabled: prefs.HealthAlertsEnabled,
		VisitReminders:      prefs.VisitReminders,
		VisitReminderHours:  prefs.VisitReminderHours,
		VaccineReminders:    prefs.VaccineReminders,
		CreatedAt:           prefs.CreatedAt,
		UpdatedAt:           prefs.UpdatedAt,
	}, nil
}

// CreateDefaultPreferences creates default notification preferences for a user
func (s *Service) CreateDefaultPreferences(
	ctx context.Context,
	userID uuid.UUID,
) (*UserPreferences, error) {
	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to verify user exists")
	}

	prefs := &model.NotificationPreference{
		ID:                  uuid.New(),
		UserID:              userID,
		SMSEnabled:          true,
		PushEnabled:         true,
		EmailEnabled:        false,  // Email disabled by default as not primary communication channel
		QuietHoursEnabled:   false,
		QuietHoursStart:     "22:00",
		QuietHoursEnd:       "07:00",
		PreferredChannel:    "sms", // SMS is default for Sierra Leone context
		HealthAlertsEnabled: true,
		VisitReminders:      true,
		VisitReminderHours:  24, // Remind 24 hours before visit
		VaccineReminders:    true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.notifPrefsRepo.Create(ctx, prefs); err != nil {
		return nil, errorx.Wrap(err, "failed to create default notification preferences")
	}

	s.log.Info("Created default notification preferences", logger.Fields{
		"user_id": userID.String(),
	})

	return &UserPreferences{
		ID:                  prefs.ID,
		UserID:              prefs.UserID,
		SMSEnabled:          prefs.SMSEnabled,
		PushEnabled:         prefs.PushEnabled,
		EmailEnabled:        prefs.EmailEnabled,
		QuietHoursEnabled:   prefs.QuietHoursEnabled,
		QuietHoursStart:     prefs.QuietHoursStart,
		QuietHoursEnd:       prefs.QuietHoursEnd,
		PreferredChannel:    prefs.PreferredChannel,
		HealthAlertsEnabled: prefs.HealthAlertsEnabled,
		VisitReminders:      prefs.VisitReminders,
		VisitReminderHours:  prefs.VisitReminderHours,
		VaccineReminders:    prefs.VaccineReminders,
		CreatedAt:           prefs.CreatedAt,
		UpdatedAt:           prefs.UpdatedAt,
	}, nil
}

// UpdateUserPreferences updates notification preferences for a user
func (s *Service) UpdateUserPreferences(
	ctx context.Context,
	userID uuid.UUID,
	preferences *UserPreferences,
) (*UserPreferences, error) {
	// Retrieve existing preferences
	existingPrefs, err := s.notifPrefsRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If preferences don't exist, create them first
		if errorx.IsNotFound(err) {
			_, err = s.CreateDefaultPreferences(ctx, userID)
			if err != nil {
				return nil, err
			}
			existingPrefs, err = s.notifPrefsRepo.GetByUserID(ctx, userID)
			if err != nil {
				return nil, errorx.Wrap(err, "failed to get newly created preferences")
			}
		} else {
			return nil, errorx.Wrap(err, "failed to get notification preferences")
		}
	}

	// Update the preferences
	existingPrefs.SMSEnabled = preferences.SMSEnabled
	existingPrefs.PushEnabled = preferences.PushEnabled
	existingPrefs.EmailEnabled = preferences.EmailEnabled
	existingPrefs.QuietHoursEnabled = preferences.QuietHoursEnabled
	existingPrefs.PreferredChannel = preferences.PreferredChannel
	existingPrefs.HealthAlertsEnabled = preferences.HealthAlertsEnabled
	existingPrefs.VisitReminders = preferences.VisitReminders
	existingPrefs.VaccineReminders = preferences.VaccineReminders
	existingPrefs.UpdatedAt = time.Now()

	// Update optional fields only if provided
	if preferences.QuietHoursStart != "" {
		existingPrefs.QuietHoursStart = preferences.QuietHoursStart
	}
	if preferences.QuietHoursEnd != "" {
		existingPrefs.QuietHoursEnd = preferences.QuietHoursEnd
	}
	if preferences.VisitReminderHours > 0 {
		existingPrefs.VisitReminderHours = preferences.VisitReminderHours
	}

	if err := s.notifPrefsRepo.Update(ctx, existingPrefs); err != nil {
		return nil, errorx.Wrap(err, "failed to update notification preferences")
	}

	s.log.Info("Updated notification preferences", logger.Fields{
		"user_id": userID.String(),
	})

	return &UserPreferences{
		ID:                  existingPrefs.ID,
		UserID:              existingPrefs.UserID,
		SMSEnabled:          existingPrefs.SMSEnabled,
		PushEnabled:         existingPrefs.PushEnabled,
		EmailEnabled:        existingPrefs.EmailEnabled,
		QuietHoursEnabled:   existingPrefs.QuietHoursEnabled,
		QuietHoursStart:     existingPrefs.QuietHoursStart,
		QuietHoursEnd:       existingPrefs.QuietHoursEnd,
		PreferredChannel:    existingPrefs.PreferredChannel,
		HealthAlertsEnabled: existingPrefs.HealthAlertsEnabled,
		VisitReminders:      existingPrefs.VisitReminders,
		VisitReminderHours:  existingPrefs.VisitReminderHours,
		VaccineReminders:    existingPrefs.VaccineReminders,
		CreatedAt:           existingPrefs.CreatedAt,
		UpdatedAt:           existingPrefs.UpdatedAt,
	}, nil
}

// IsNotificationEnabled checks if a specific type of notification is enabled for a user
func (s *Service) IsNotificationEnabled(
	ctx context.Context,
	userID uuid.UUID,
	notificationType NotificationType,
) (bool, error) {
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		return false, err
	}

	// Check if any notifications are enabled
	if !prefs.SMSEnabled && !prefs.PushEnabled && !prefs.EmailEnabled {
		return false, nil
	}

	// Check quiet hours
	if prefs.QuietHoursEnabled && notificationType != TypeEmergency {
		if s.isCurrentlyInQuietHours(prefs.QuietHoursStart, prefs.QuietHoursEnd) {
			return false, nil
		}
	}

	// Check specific notification type
	switch notificationType {
	case TypeHealthAlert:
		return prefs.HealthAlertsEnabled, nil
	case TypeVisitReminder:
		return prefs.VisitReminders, nil
	case TypeVaccineReminder:
		return prefs.VaccineReminders, nil
	case TypeEmergency:
		// Emergency notifications are always enabled
		return true, nil
	default:
		// For other notification types, default to enabled
		return true, nil
	}
}

// GetPreferredChannel gets the preferred notification channel for a user
func (s *Service) GetPreferredChannel(
	ctx context.Context,
	userID uuid.UUID,
) (string, error) {
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		return "", err
	}

	// Return preferred channel only if it's enabled
	switch prefs.PreferredChannel {
	case "sms":
		if prefs.SMSEnabled {
			return "sms", nil
		}
	case "push":
		if prefs.PushEnabled {
			return "push", nil
		}
	case "email":
		if prefs.EmailEnabled {
			return "email", nil
		}
	}

	// If preferred channel is disabled, return the first enabled channel
	if prefs.SMSEnabled {
		return "sms", nil
	}
	if prefs.PushEnabled {
		return "push", nil
	}
	if prefs.EmailEnabled {
		return "email", nil
	}

	// No channels enabled
	return "", errorx.New(errorx.NotFound, "no notification channels enabled for user")
}

// DisableAllNotifications disables all notifications for a user
func (s *Service) DisableAllNotifications(
	ctx context.Context,
	userID uuid.UUID,
) error {
	prefs, err := s.notifPrefsRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If preferences don't exist, create default preferences with notifications disabled
		if errorx.IsNotFound(err) {
			prefs = &model.NotificationPreference{
				ID:                  uuid.New(),
				UserID:              userID,
				SMSEnabled:          false,
				PushEnabled:         false,
				EmailEnabled:        false,
				QuietHoursEnabled:   false,
				PreferredChannel:    "sms",
				HealthAlertsEnabled: false,
				VisitReminders:      false,
				VaccineReminders:    false,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			}
			return s.notifPrefsRepo.Create(ctx, prefs)
		}
		return errorx.Wrap(err, "failed to get notification preferences")
	}

	// Disable all notifications
	prefs.SMSEnabled = false
	prefs.PushEnabled = false
	prefs.EmailEnabled = false
	prefs.HealthAlertsEnabled = false
	prefs.VisitReminders = false
	prefs.VaccineReminders = false
	prefs.UpdatedAt = time.Now()

	if err := s.notifPrefsRepo.Update(ctx, prefs); err != nil {
		return errorx.Wrap(err, "failed to update notification preferences")
	}

	s.log.Info("Disabled all notifications", logger.Fields{
		"user_id": userID.String(),
	})

	return nil
}

// isCurrentlyInQuietHours checks if the current time is within quiet hours
func (s *Service) isCurrentlyInQuietHours(start, end string) bool {
	now := time.Now()
	nowTimeStr := now.Format("15:04")

	// Parse time strings into comparable format
	startParts := make([]int, 2)
	endParts := make([]int, 2)
	nowParts := make([]int, 2)

	_, err := fmt.Sscanf(start, "%d:%d", &startParts[0], &startParts[1])
	if err != nil {
		return false
	}
	_, err = fmt.Sscanf(end, "%d:%d", &endParts[0], &endParts[1])
	if err != nil {
		return false
	}
	_, err = fmt.Sscanf(nowTimeStr, "%d:%d", &nowParts[0], &nowParts[1])
	if err != nil {
		return false
	}

	// Convert to minutes for easier comparison
	startMinutes := startParts[0]*60 + startParts[1]
	endMinutes := endParts[0]*60 + endParts[1]
	nowMinutes := nowParts[0]*60 + nowParts[1]

	// Check if quiet hours span across midnight
	if startMinutes > endMinutes {
		// Example: 22:00 to 07:00
		return nowMinutes >= startMinutes || nowMinutes <= endMinutes
	}

	// Normal case: start time is before end time
	return nowMinutes >= startMinutes && nowMinutes <= endMinutes
}
