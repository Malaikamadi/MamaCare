package push

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

// NotificationPriority defines the priority level of a push notification
type NotificationPriority string

const (
	// PriorityNormal is for standard notifications that don't require immediate attention
	PriorityNormal NotificationPriority = "normal"
	// PriorityHigh is for important notifications that should be delivered immediately
	PriorityHigh NotificationPriority = "high"
	// PriorityCritical is for urgent notifications that require immediate user action
	PriorityCritical NotificationPriority = "critical"
)

// PushProvider defines the push notification provider
type PushProvider string

const (
	// ProviderExpo uses Expo Push Notification Service
	ProviderExpo PushProvider = "expo"
	// ProviderFirebase uses Firebase Cloud Messaging
	ProviderFirebase PushProvider = "firebase"
)

// NotificationPayload contains the content of a push notification
type NotificationPayload struct {
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Sound       string                 `json:"sound,omitempty"`
	Badge       int                    `json:"badge,omitempty"`
	ChannelID   string                 `json:"channelId,omitempty"`
	CategoryID  string                 `json:"categoryId,omitempty"`
	Priority    NotificationPriority   `json:"priority,omitempty"`
	TTL         int                    `json:"ttl,omitempty"`          // Time to live in seconds
	Expiration  int                    `json:"expiration,omitempty"`   // Timestamp for when the notification expires
	InteractID  string                 `json:"interactId,omitempty"`   // ID for interacting with the notification
	IsBackground bool                  `json:"isBackground,omitempty"` // Whether the notification should be processed in the background
}

// PushResult represents the result of a push notification delivery attempt
type PushResult struct {
	ID          string    `json:"id"`          // ID of the notification
	Status      string    `json:"status"`      // Status of the delivery attempt (success, error)
	Provider    string    `json:"provider"`    // Provider used to send the notification
	StatusCode  int       `json:"statusCode"`  // HTTP status code from the provider
	Error       string    `json:"error"`       // Error message, if any
	ErrorCode   string    `json:"errorCode"`   // Error code from the provider
	SentTime    time.Time `json:"sentTime"`    // Time when the notification was sent
	DeliveredAt time.Time `json:"deliveredAt"` // Time when the notification was delivered (if known)
}

// Service handles push notification delivery
type Service struct {
	deviceTokenRepo repository.DeviceTokenRepository
	notifHistoryRepo repository.NotificationHistoryRepository
	notifPrefsRepo repository.NotificationPreferenceRepository
	log             logger.Logger
	expoAPIKey      string
	firebaseConfig  string
	defaultProvider PushProvider
}

// NewService creates a new push notification service
func NewService(
	deviceTokenRepo repository.DeviceTokenRepository,
	notifHistoryRepo repository.NotificationHistoryRepository,
	notifPrefsRepo repository.NotificationPreferenceRepository,
	log logger.Logger,
	expoAPIKey string,
	firebaseConfig string,
) *Service {
	return &Service{
		deviceTokenRepo: deviceTokenRepo,
		notifHistoryRepo: notifHistoryRepo,
		notifPrefsRepo: notifPrefsRepo,
		log:             log,
		expoAPIKey:      expoAPIKey,
		firebaseConfig:  firebaseConfig,
		defaultProvider: ProviderExpo, // Expo is the default provider for MamaCare
	}
}

// SendToUser sends a push notification to a specific user
func (s *Service) SendToUser(
	ctx context.Context,
	userID uuid.UUID,
	payload *NotificationPayload,
	provider PushProvider,
) (*PushResult, error) {
	// Check if user has opted out of push notifications
	prefs, err := s.notifPrefsRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get notification preferences", logger.Fields{
			"user_id": userID.String(),
			"error":   err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to get notification preferences")
	}

	if !prefs.PushEnabled {
		s.log.Info("User has disabled push notifications", logger.Fields{
			"user_id": userID.String(),
		})
		return nil, errorx.New(errorx.OperationFailed, "user has disabled push notifications")
	}

	// Get the user's device tokens
	tokens, err := s.deviceTokenRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.log.Error("Failed to get device tokens", logger.Fields{
			"user_id": userID.String(),
			"error":   err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to get device tokens")
	}

	if len(tokens) == 0 {
		s.log.Info("No device tokens found for user", logger.Fields{
			"user_id": userID.String(),
		})
		return nil, errorx.New(errorx.NotFound, "no device tokens found for user")
	}

	// Use the provided provider or the default if not specified
	if provider == "" {
		provider = s.defaultProvider
	}

	// Prepare batch tokens for sending
	tokenStrings := make([]string, len(tokens))
	for i, token := range tokens {
		tokenStrings[i] = token.Token
	}

	// Send the notification
	result, err := s.sendPushNotification(ctx, tokenStrings, payload, provider)
	if err != nil {
		s.log.Error("Failed to send push notification", logger.Fields{
			"user_id":  userID.String(),
			"provider": string(provider),
			"error":    err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to send push notification")
	}

	// Record notification in history
	notification := &model.NotificationHistory{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      "push",
		Channel:   string(provider),
		Title:     payload.Title,
		Content:   payload.Body,
		Status:    result.Status,
		Error:     result.Error,
		SentAt:    result.SentTime,
		CreatedAt: time.Now(),
	}

	// Add metadata if available
	if len(payload.Data) > 0 {
		metadataBytes, err := json.Marshal(payload.Data)
		if err == nil {
			notification.Metadata = string(metadataBytes)
		}
	}

	// Save to history
	if err := s.notifHistoryRepo.Create(ctx, notification); err != nil {
		s.log.Error("Failed to save notification history", logger.Fields{
			"user_id":          userID.String(),
			"notification_id":  notification.ID.String(),
			"error":            err.Error(),
		})
		// We don't return this error since the notification was still sent
	}

	return result, nil
}

// SendToMultipleUsers sends a push notification to multiple users
func (s *Service) SendToMultipleUsers(
	ctx context.Context,
	userIDs []uuid.UUID,
	payload *NotificationPayload,
	provider PushProvider,
) (map[uuid.UUID]*PushResult, error) {
	if len(userIDs) == 0 {
		return nil, errorx.New(errorx.BadRequest, "no user IDs provided")
	}

	results := make(map[uuid.UUID]*PushResult)
	for _, userID := range userIDs {
		result, err := s.SendToUser(ctx, userID, payload, provider)
		if err != nil {
			// Log the error but continue with other users
			s.log.Error("Failed to send notification to user", logger.Fields{
				"user_id": userID.String(),
				"error":   err.Error(),
			})
			results[userID] = &PushResult{
				Status:    "error",
				Error:     err.Error(),
				Provider:  string(provider),
				SentTime:  time.Now(),
			}
		} else {
			results[userID] = result
		}
	}

	return results, nil
}

// SendToTopic sends a push notification to all users subscribed to a topic
func (s *Service) SendToTopic(
	ctx context.Context,
	topic string,
	payload *NotificationPayload,
	provider PushProvider,
) (*PushResult, error) {
	if topic == "" {
		return nil, errorx.New(errorx.BadRequest, "topic cannot be empty")
	}

	// Use the provided provider or the default if not specified
	if provider == "" {
		provider = s.defaultProvider
	}

	// For Firebase, we can send directly to topics
	if provider == ProviderFirebase {
		return s.sendFirebaseTopicNotification(ctx, topic, payload)
	}

	// For Expo, we need to get all device tokens subscribed to the topic
	tokens, err := s.deviceTokenRepo.GetByTopic(ctx, topic)
	if err != nil {
		s.log.Error("Failed to get device tokens for topic", logger.Fields{
			"topic": topic,
			"error": err.Error(),
		})
		return nil, errorx.Wrap(err, "failed to get device tokens for topic")
	}

	if len(tokens) == 0 {
		s.log.Info("No device tokens found for topic", logger.Fields{
			"topic": topic,
		})
		return nil, errorx.New(errorx.NotFound, "no device tokens found for topic")
	}

	// Extract token strings
	tokenStrings := make([]string, len(tokens))
	for i, token := range tokens {
		tokenStrings[i] = token.Token
	}

	// Send the notification
	return s.sendPushNotification(ctx, tokenStrings, payload, provider)
}

// sendPushNotification sends the notification via the specified provider
func (s *Service) sendPushNotification(
	ctx context.Context,
	tokens []string,
	payload *NotificationPayload,
	provider PushProvider,
) (*PushResult, error) {
	switch provider {
	case ProviderExpo:
		return s.sendExpoPushNotification(ctx, tokens, payload)
	case ProviderFirebase:
		return s.sendFirebasePushNotification(ctx, tokens, payload)
	default:
		return nil, errorx.New(errorx.BadRequest, fmt.Sprintf("unsupported provider: %s", provider))
	}
}

// sendExpoPushNotification sends a push notification via Expo Push Notification Service
func (s *Service) sendExpoPushNotification(
	ctx context.Context,
	tokens []string,
	payload *NotificationPayload,
) (*PushResult, error) {
	// TODO: Implement actual Expo Push API integration
	// For MVP, we'll just simulate a successful send

	s.log.Info("Sending Expo push notification", logger.Fields{
		"token_count": len(tokens),
		"title":       payload.Title,
	})

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	// In a real implementation, we would:
	// 1. Build the Expo API request
	// 2. Make HTTP request to Expo Push API
	// 3. Parse response and handle errors
	// 4. Return appropriate result

	// Simulate successful result
	result := &PushResult{
		ID:         uuid.New().String(),
		Status:     "success",
		Provider:   string(ProviderExpo),
		StatusCode: 200,
		SentTime:   time.Now(),
	}

	return result, nil
}

// sendFirebasePushNotification sends a push notification via Firebase Cloud Messaging
func (s *Service) sendFirebasePushNotification(
	ctx context.Context,
	tokens []string,
	payload *NotificationPayload,
) (*PushResult, error) {
	// TODO: Implement actual Firebase Cloud Messaging integration
	// For MVP, we'll just simulate a successful send

	s.log.Info("Sending Firebase push notification", logger.Fields{
		"token_count": len(tokens),
		"title":       payload.Title,
	})

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	// In a real implementation, we would:
	// 1. Initialize Firebase Admin SDK with provided config
	// 2. Build the FCM message
	// 3. Send the message to FCM
	// 4. Parse response and handle errors
	// 5. Return appropriate result

	// Simulate successful result
	result := &PushResult{
		ID:         uuid.New().String(),
		Status:     "success",
		Provider:   string(ProviderFirebase),
		StatusCode: 200,
		SentTime:   time.Now(),
	}

	return result, nil
}

// sendFirebaseTopicNotification sends a notification to a Firebase topic
func (s *Service) sendFirebaseTopicNotification(
	ctx context.Context,
	topic string,
	payload *NotificationPayload,
) (*PushResult, error) {
	// TODO: Implement actual Firebase Cloud Messaging topic integration
	// For MVP, we'll just simulate a successful send

	s.log.Info("Sending Firebase topic notification", logger.Fields{
		"topic": topic,
		"title": payload.Title,
	})

	// Simulate API delay
	time.Sleep(100 * time.Millisecond)

	// Simulate successful result
	result := &PushResult{
		ID:         uuid.New().String(),
		Status:     "success",
		Provider:   string(ProviderFirebase),
		StatusCode: 200,
		SentTime:   time.Now(),
	}

	return result, nil
}

// ValidateExpoToken validates if an Expo push token is correctly formatted
func (s *Service) ValidateExpoToken(token string) bool {
	// Expo push tokens start with ExponentPushToken[ and end with ]
	// And must be at least 20 characters long
	if len(token) < 20 {
		return false
	}

	return token[:17] == "ExponentPushToken[" && token[len(token)-1:] == "]"
}

// ValidateFirebaseToken validates if a Firebase token is correctly formatted
func (s *Service) ValidateFirebaseToken(token string) bool {
	// Firebase tokens are long random strings
	// Must be at least 140 characters long
	return len(token) >= 140
}
