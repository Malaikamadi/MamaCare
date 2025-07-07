package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// SMSPriority defines the priority level of an SMS message
type SMSPriority string

const (
	// PriorityNormal is for standard messages
	PriorityNormal SMSPriority = "normal"
	// PriorityHigh is for important messages
	PriorityHigh SMSPriority = "high"
	// PriorityCritical is for urgent messages
	PriorityCritical SMSPriority = "critical"
)

// SMSProvider defines the SMS sending provider
type SMSProvider string

const (
	// ProviderTwilio uses Twilio for SMS delivery
	ProviderTwilio SMSProvider = "twilio"
	// ProviderAfricasTalking uses Africa's Talking for SMS delivery
	ProviderAfricasTalking SMSProvider = "africas_talking"
)

// MessagePayload contains the content of an SMS message
type MessagePayload struct {
	Body        string      `json:"body"`                  // Message content
	Data        interface{} `json:"data,omitempty"`        // Additional data for templating
	Priority    SMSPriority `json:"priority,omitempty"`    // Message priority
	SenderID    string      `json:"senderId,omitempty"`    // Sender ID or phone number
	Template    string      `json:"template,omitempty"`    // Template ID for templated messages
	MediaURL    string      `json:"mediaUrl,omitempty"`    // URL for MMS media
	ScheduleFor *time.Time  `json:"scheduleFor,omitempty"` // Future time to send message
}

// SMSResult represents the result of an SMS delivery attempt
type SMSResult struct {
	ID          string    `json:"id"`          // SMS provider message ID
	Status      string    `json:"status"`      // Status of the delivery attempt (queued, sent, delivered, failed)
	Provider    string    `json:"provider"`    // Provider used to send the message
	StatusCode  int       `json:"statusCode"`  // HTTP status code from the provider
	Error       string    `json:"error"`       // Error message if any
	ErrorCode   string    `json:"errorCode"`   // Error code from the provider
	SentTime    time.Time `json:"sentTime"`    // Time when the message was sent
	DeliveredAt time.Time `json:"deliveredAt"` // Time when the message was delivered (if available)
	Cost        float64   `json:"cost"`        // Cost of sending the message (if available)
	Currency    string    `json:"currency"`    // Currency of the cost (if available)
}

// Service handles SMS message delivery
type Service struct {
	userRepo         repository.UserRepository
	notifHistoryRepo repository.NotificationHistoryRepository
	notifPrefsRepo   repository.NotificationPreferenceRepository
	log              logger.Logger
	twilioAccountSID string
	twilioAuthToken  string
	twilioPhoneNum   string
	atAPIKey         string
	atUsername       string
	defaultProvider  SMSProvider
}

// NewService creates a new SMS notification service
func NewService(
	userRepo repository.UserRepository,
	notifHistoryRepo repository.NotificationHistoryRepository,
	notifPrefsRepo repository.NotificationPreferenceRepository,
	log logger.Logger,
	twilioAccountSID string,
	twilioAuthToken string,
	twilioPhoneNum string,
	atAPIKey string,
	atUsername string,
) *Service {
	return &Service{
		userRepo:         userRepo,
		notifHistoryRepo: notifHistoryRepo,
		notifPrefsRepo:   notifPrefsRepo,
		log:              log,
		twilioAccountSID: twilioAccountSID,
		twilioAuthToken:  twilioAuthToken,
		twilioPhoneNum:   twilioPhoneNum,
		atAPIKey:         atAPIKey,
		atUsername:       atUsername,
		defaultProvider:  ProviderTwilio, // Twilio is the default provider for MamaCare
	}
}

// SendToUser sends an SMS message to a specific user
func (s *Service) SendToUser(
	ctx context.Context,
	userID uuid.UUID,
	payload *MessagePayload,
	provider SMSProvider,
) (*SMSResult, error) {
	// Check if user has opted out of SMS notifications
	prefs, err := s.notifPrefsRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get notification preferences", logger.Fields{
			"user_id": userID.String(),
			"error":   err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to get notification preferences")
	}

	if !prefs.SMSEnabled {
		s.log.Info("User has disabled SMS notifications", logger.Fields{
			"user_id": userID.String(),
		})
		return nil, errorx.New(errorx.OperationFailed, "user has disabled SMS notifications")
	}

	// Get the user's phone number
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get user", logger.Fields{
			"user_id": userID.String(),
			"error":   err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to get user details")
	}

	if user.PhoneNumber == "" {
		s.log.Info("User has no phone number", logger.Fields{
			"user_id": userID.String(),
		})
		return nil, errorx.New(errorx.NotFound, "user has no phone number")
	}

	// Use the provided provider or the default if not specified
	if provider == "" {
		provider = s.defaultProvider
	}

	// Check if message needs to be scheduled for future delivery
	if payload.ScheduleFor != nil && payload.ScheduleFor.After(time.Now()) {
		return s.scheduleMessage(ctx, user.PhoneNumber, payload, provider, userID)
	}

	// Send the message immediately
	result, err := s.sendSMS(ctx, user.PhoneNumber, payload, provider)
	if err != nil {
		s.log.Error("Failed to send SMS", logger.Fields{
			"user_id":  userID.String(),
			"provider": string(provider),
			"error":    err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to send SMS")
	}

	// Record notification in history
	notification := &model.NotificationHistory{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "sms",
		Channel:   string(provider),
		Content:   payload.Body,
		Status:    result.Status,
		Error:     result.Error,
		SentAt:    result.SentTime,
		CreatedAt: time.Now(),
	}

	// Add metadata if available
	if payload.Data != nil {
		metadataBytes, err := json.Marshal(payload.Data)
		if err == nil {
			notification.Metadata = string(metadataBytes)
		}
	}

	// Save to history
	if err := s.notifHistoryRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to save notification history", logger.Fields{
			"user_id":         userID.String(),
			"notification_id": notification.ID.String(),
			"error":           err.Error(),
		})
		// We don't return this error since the SMS was still sent
	}

	return result, nil
}

// SendToMultipleUsers sends an SMS message to multiple users
func (s *Service) SendToMultipleUsers(
	ctx context.Context,
	userIDs []uuid.UUID,
	payload *MessagePayload,
	provider SMSProvider,
) (map[uuid.UUID]*SMSResult, error) {
	if len(userIDs) == 0 {
		return nil, errorx.New(errorx.BadRequest, "no user IDs provided")
	}

	results := make(map[uuid.UUID]*SMSResult)
	for _, userID := range userIDs {
		result, err := s.SendToUser(ctx, userID, payload, provider)
		if err != nil {
			// Log the error but continue with other users
			s.log.Error("Failed to send SMS to user", logger.Fields{
				"user_id": userID.String(),
				"error":   err.Error(),
			})
			results[userID] = &SMSResult{
				Status:   "error",
				Error:    err.Error(),
				Provider: string(provider),
				SentTime: time.Now(),
			}
		} else {
			results[userID] = result
		}
	}

	return results, nil
}

// SendToNumber sends an SMS message directly to a phone number
func (s *Service) SendToNumber(
	ctx context.Context,
	phoneNumber string,
	payload *MessagePayload,
	provider SMSProvider,
) (*SMSResult, error) {
	if phoneNumber == "" {
		return nil, errorx.New(errorx.BadRequest, "phone number cannot be empty")
	}

	// Use the provided provider or the default if not specified
	if provider == "" {
		provider = s.defaultProvider
	}

	// Send the message
	result, err := s.sendSMS(ctx, phoneNumber, payload, provider)
	if err != nil {
		s.log.Error("Failed to send SMS", logger.Fields{
			"phone":    phoneNumber,
			"provider": string(provider),
			"error":    err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to send SMS")
	}

	// Record in history without user ID
	notification := &model.NotificationHistory{
		ID:        uuid.New(),
		Type:      "sms",
		Channel:   string(provider),
		Recipient: phoneNumber,
		Content:   payload.Body,
		Status:    result.Status,
		Error:     result.Error,
		SentAt:    result.SentTime,
		CreatedAt: time.Now(),
	}

	// Add metadata if available
	if payload.Data != nil {
		metadataBytes, err := json.Marshal(payload.Data)
		if err == nil {
			notification.Metadata = string(metadataBytes)
		}
	}

	// Save to history
	if err := s.notifHistoryRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to save notification history", logger.Fields{
			"phone":          phoneNumber,
			"notification_id": notification.ID.String(),
			"error":           err.Error(),
		})
		// We don't return this error since the SMS was still sent
	}

	return result, nil
}

// scheduleMessage schedules an SMS message for future delivery
func (s *Service) scheduleMessage(
	ctx context.Context,
	phoneNumber string,
	payload *MessagePayload,
	provider SMSProvider,
	userID uuid.UUID,
) (*SMSResult, error) {
	// For MVP, we'll create a placeholder for scheduled messages
	// In a real implementation, we would create a database record and use a job queue

	s.log.Info("Scheduling SMS message", logger.Fields{
		"phone":           phoneNumber,
		"scheduled_time": payload.ScheduleFor.Format(time.RFC3339),
		"provider":       string(provider),
	})

	// Create a successful schedule result
	result := &SMSResult{
		ID:       uuid.New().String(),
		Status:   "scheduled",
		Provider: string(provider),
		SentTime: *payload.ScheduleFor,
	}

	// Record scheduled notification in history
	notification := &model.NotificationHistory{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "sms",
		Channel:   string(provider),
		Recipient: phoneNumber,
		Content:   payload.Body,
		Status:    "scheduled",
		SentAt:    *payload.ScheduleFor,
		CreatedAt: time.Now(),
	}

	// Add metadata if available
	if payload.Data != nil {
		metadataBytes, err := json.Marshal(payload.Data)
		if err == nil {
			notification.Metadata = string(metadataBytes)
		}
	}

	// Save to history
	if err := s.notifHistoryRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to save scheduled notification history", logger.Fields{
			"notification_id": notification.ID.String(),
			"error":           err.Error(),
		})
		// We don't return this error since the schedule was still created
	}

	return result, nil
}

// sendSMS sends the SMS message via the specified provider
func (s *Service) sendSMS(
	ctx context.Context,
	phoneNumber string,
	payload *MessagePayload,
	provider SMSProvider,
) (*SMSResult, error) {
	switch provider {
	case ProviderTwilio:
		return s.sendTwilioSMS(ctx, phoneNumber, payload)
	case ProviderAfricasTalking:
		return s.sendAfricasTalkingSMS(ctx, phoneNumber, payload)
	default:
		return nil, errorx.New(errorx.BadRequest, fmt.Sprintf("unsupported provider: %s", provider))
	}
}

// sendTwilioSMS sends an SMS message via Twilio
func (s *Service) sendTwilioSMS(
	ctx context.Context,
	phoneNumber string,
	payload *MessagePayload,
) (*SMSResult, error) {
	// TODO: Implement actual Twilio API integration
	// For MVP, we'll just simulate a successful send

	s.log.Info("Sending Twilio SMS", logger.Fields{
		"phone": phoneNumber,
		"body_length": len(payload.Body),
	})

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	// In a real implementation, we would:
	// 1. Build the Twilio API request
	// 2. Make HTTP request to Twilio API
	// 3. Parse response and handle errors
	// 4. Return appropriate result

	// Simulate successful result
	result := &SMSResult{
		ID:         "SM" + uuid.New().String()[0:12],
		Status:     "sent",
		Provider:   string(ProviderTwilio),
		StatusCode: 201,
		SentTime:   time.Now(),
		Cost:       0.05,
		Currency:   "USD",
	}

	return result, nil
}

// sendAfricasTalkingSMS sends an SMS message via Africa's Talking
func (s *Service) sendAfricasTalkingSMS(
	ctx context.Context,
	phoneNumber string,
	payload *MessagePayload,
) (*SMSResult, error) {
	// TODO: Implement actual Africa's Talking API integration
	// For MVP, we'll just simulate a successful send

	s.log.Info("Sending Africa's Talking SMS", logger.Fields{
		"phone": phoneNumber,
		"body_length": len(payload.Body),
	})

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	// In a real implementation, we would:
	// 1. Build the Africa's Talking API request
	// 2. Make HTTP request to Africa's Talking API
	// 3. Parse response and handle errors
	// 4. Return appropriate result

	// Simulate successful result
	result := &SMSResult{
		ID:         "AT" + uuid.New().String()[0:12],
		Status:     "sent",
		Provider:   string(ProviderAfricasTalking),
		StatusCode: 201,
		SentTime:   time.Now(),
		Cost:       0.03,
		Currency:   "USD",
	}

	return result, nil
}

// ValidateE164Format checks if a phone number is in E.164 format
func (s *Service) ValidateE164Format(phoneNumber string) bool {
	// E.164 format requires a + prefix followed by country code and number
	// Must be at least 8 characters (including the +)
	if len(phoneNumber) < 8 || phoneNumber[0] != '+' {
		return false
	}

	// Check if the rest is only digits
	for _, c := range phoneNumber[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}

	return true
}
