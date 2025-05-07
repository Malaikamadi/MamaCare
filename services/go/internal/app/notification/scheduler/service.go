package scheduler

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/app/notification/push"
	"github.com/mamacare/services/internal/app/notification/sms"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// NotificationType defines the type of notification to be scheduled
type NotificationType string

const (
	// TypeSMS is for SMS messages
	TypeSMS NotificationType = "sms"
	// TypePush is for push notifications
	TypePush NotificationType = "push"
)

// NotificationStatus represents the status of a scheduled notification
type NotificationStatus string

const (
	// StatusPending indicates the notification is waiting to be sent
	StatusPending NotificationStatus = "pending"
	// StatusProcessing indicates the notification is being processed
	StatusProcessing NotificationStatus = "processing"
	// StatusSent indicates the notification was successfully sent
	StatusSent NotificationStatus = "sent"
	// StatusFailed indicates the notification failed to send
	StatusFailed NotificationStatus = "failed"
	// StatusCancelled indicates the notification was cancelled before sending
	StatusCancelled NotificationStatus = "cancelled"
)

// ScheduledNotification represents a notification scheduled for future delivery
type ScheduledNotification struct {
	ID         uuid.UUID         `json:"id"`
	UserID     uuid.UUID         `json:"user_id,omitempty"`
	Recipient  string            `json:"recipient,omitempty"` // Phone number or device token
	Type       NotificationType  `json:"type"`
	Status     NotificationStatus `json:"status"`
	Provider   string            `json:"provider"`
	Content    interface{}       `json:"content"`    // SMS message or push payload
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ScheduleAt time.Time         `json:"schedule_at"`
	SentAt     *time.Time        `json:"sent_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Error      string            `json:"error,omitempty"`
}

// Service handles scheduling and processing of notifications
type Service struct {
	scheduledNotifRepo repository.ScheduledNotificationRepository
	notifHistoryRepo   repository.NotificationHistoryRepository
	pushService        *push.Service
	smsService         *sms.Service
	log                logger.Logger
	processingInterval time.Duration
	batchSize          int
	processing         bool
	mu                 sync.Mutex
}

// NewService creates a new notification scheduler service
func NewService(
	scheduledNotifRepo repository.ScheduledNotificationRepository,
	notifHistoryRepo repository.NotificationHistoryRepository,
	pushService *push.Service,
	smsService *sms.Service,
	log logger.Logger,
	processingInterval time.Duration,
) *Service {
	if processingInterval == 0 {
		processingInterval = 1 * time.Minute
	}

	return &Service{
		scheduledNotifRepo: scheduledNotifRepo,
		notifHistoryRepo:   notifHistoryRepo,
		pushService:        pushService,
		smsService:         smsService,
		log:                log,
		processingInterval: processingInterval,
		batchSize:          50, // Process 50 notifications at once
		processing:         false,
	}
}

// ScheduleSMS schedules an SMS notification for future delivery
func (s *Service) ScheduleSMS(
	ctx context.Context,
	userID uuid.UUID,
	phoneNumber string,
	content *sms.MessagePayload,
	provider sms.SMSProvider,
	scheduleAt time.Time,
) (*ScheduledNotification, error) {
	if userID == uuid.Nil && phoneNumber == "" {
		return nil, errorx.New(errorx.BadRequest, "either user ID or phone number must be provided")
	}

	if content == nil || content.Body == "" {
		return nil, errorx.New(errorx.BadRequest, "message content is required")
	}

	if scheduleAt.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "schedule time must be in the future")
	}

	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to marshal SMS content")
	}

	notification := &model.ScheduledNotification{
		ID:         uuid.New(),
		UserID:     userID,
		Recipient:  phoneNumber,
		Type:       string(TypeSMS),
		Status:     string(StatusPending),
		Provider:   string(provider),
		Content:    string(contentBytes),
		ScheduleAt: scheduleAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.scheduledNotifRepo.Create(ctx, notification); err != nil {
		return nil, errorx.Wrap(err, "failed to create scheduled notification")
	}

	s.log.Info("Scheduled SMS notification", logger.Fields{
		"notification_id": notification.ID.String(),
		"schedule_at":     scheduleAt.Format(time.RFC3339),
	})

	// Convert to response format
	var metadata map[string]interface{}
	if content.Data != nil {
		metadata = map[string]interface{}{
			"data": content.Data,
		}
	}

	scheduledNotif := &ScheduledNotification{
		ID:         notification.ID,
		UserID:     notification.UserID,
		Recipient:  notification.Recipient,
		Type:       TypeSMS,
		Status:     StatusPending,
		Provider:   notification.Provider,
		Content:    content,
		Metadata:   metadata,
		ScheduleAt: notification.ScheduleAt,
		CreatedAt:  notification.CreatedAt,
		UpdatedAt:  notification.UpdatedAt,
	}

	return scheduledNotif, nil
}

// SchedulePush schedules a push notification for future delivery
func (s *Service) SchedulePush(
	ctx context.Context,
	userID uuid.UUID,
	content *push.NotificationPayload,
	provider push.PushProvider,
	scheduleAt time.Time,
) (*ScheduledNotification, error) {
	if userID == uuid.Nil {
		return nil, errorx.New(errorx.BadRequest, "user ID must be provided")
	}

	if content == nil || content.Title == "" || content.Body == "" {
		return nil, errorx.New(errorx.BadRequest, "notification content must include title and body")
	}

	if scheduleAt.Before(time.Now()) {
		return nil, errorx.New(errorx.BadRequest, "schedule time must be in the future")
	}

	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to marshal push content")
	}

	notification := &model.ScheduledNotification{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       string(TypePush),
		Status:     string(StatusPending),
		Provider:   string(provider),
		Content:    string(contentBytes),
		ScheduleAt: scheduleAt,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.scheduledNotifRepo.Create(ctx, notification); err != nil {
		return nil, errorx.Wrap(err, "failed to create scheduled notification")
	}

	s.log.Info("Scheduled push notification", logger.Fields{
		"notification_id": notification.ID.String(),
		"schedule_at":     scheduleAt.Format(time.RFC3339),
	})

	// Convert to response format
	var metadata map[string]interface{}
	if content.Data != nil {
		metadata = map[string]interface{}{
			"data": content.Data,
		}
	}

	scheduledNotif := &ScheduledNotification{
		ID:         notification.ID,
		UserID:     notification.UserID,
		Type:       TypePush,
		Status:     StatusPending,
		Provider:   notification.Provider,
		Content:    content,
		Metadata:   metadata,
		ScheduleAt: notification.ScheduleAt,
		CreatedAt:  notification.CreatedAt,
		UpdatedAt:  notification.UpdatedAt,
	}

	return scheduledNotif, nil
}

// CancelScheduledNotification cancels a scheduled notification
func (s *Service) CancelScheduledNotification(
	ctx context.Context,
	notificationID uuid.UUID,
) error {
	notification, err := s.scheduledNotifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return errorx.Wrap(err, "failed to get scheduled notification")
	}

	if notification.Status != string(StatusPending) {
		return errorx.New(errorx.BadRequest, "only pending notifications can be cancelled")
	}

	notification.Status = string(StatusCancelled)
	notification.UpdatedAt = time.Now()

	if err := s.scheduledNotifRepo.Update(ctx, notification); err != nil {
		return errorx.Wrap(err, "failed to update scheduled notification")
	}

	s.log.Info("Cancelled scheduled notification", logger.Fields{
		"notification_id": notification.ID.String(),
	})

	return nil
}

// GetScheduledNotification retrieves a scheduled notification by ID
func (s *Service) GetScheduledNotification(
	ctx context.Context,
	notificationID uuid.UUID,
) (*ScheduledNotification, error) {
	notification, err := s.scheduledNotifRepo.GetByID(ctx, notificationID)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to get scheduled notification")
	}

	return s.convertToScheduledNotification(notification)
}

// GetScheduledNotificationsForUser retrieves all scheduled notifications for a user
func (s *Service) GetScheduledNotificationsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]*ScheduledNotification, error) {
	notifications, err := s.scheduledNotifRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errorx.Wrap(err, "failed to get scheduled notifications for user")
	}

	result := make([]*ScheduledNotification, 0, len(notifications))
	for _, notification := range notifications {
		scheduledNotif, err := s.convertToScheduledNotification(notification)
		if err != nil {
			s.log.Error("Failed to convert scheduled notification", logger.Fields{
				"notification_id": notification.ID.String(),
				"error":           err.Error(),
			})
			continue
		}
		result = append(result, scheduledNotif)
	}

	return result, nil
}

// StartProcessing starts the scheduler processing loop
func (s *Service) StartProcessing(ctx context.Context) error {
	s.mu.Lock()
	if s.processing {
		s.mu.Unlock()
		return errorx.New(errorx.OperationFailed, "notification processing is already running")
	}
	s.processing = true
	s.mu.Unlock()

	s.log.Info("Starting scheduled notification processing", logger.Fields{
		"interval": s.processingInterval.String(),
		"batch_size": s.batchSize,
	})

	ticker := time.NewTicker(s.processingInterval)
	defer ticker.Stop()

	// Process immediately on start
	s.processScheduledNotifications(ctx)

	for {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.processing = false
			s.mu.Unlock()
			s.log.Info("Stopping scheduled notification processing", nil)
			return nil
		case <-ticker.C:
			s.processScheduledNotifications(ctx)
		}
	}
}

// StopProcessing stops the scheduler processing loop
func (s *Service) StopProcessing() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processing = false
	s.log.Info("Notification processing stopped", nil)
}

// IsProcessing returns whether the scheduler is currently processing
func (s *Service) IsProcessing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.processing
}

// processScheduledNotifications processes pending scheduled notifications
func (s *Service) processScheduledNotifications(ctx context.Context) {
	// Get notifications due for processing
	notifications, err := s.scheduledNotifRepo.GetDueNotifications(ctx, time.Now(), s.batchSize)
	if err != nil {
		s.log.Error("Failed to get due notifications", logger.Fields{
			"error": err.Error(),
		})
		return
	}

	if len(notifications) == 0 {
		return
	}

	s.log.Info("Processing scheduled notifications", logger.Fields{
		"count": len(notifications),
	})

	// Process each notification
	for _, notification := range notifications {
		// Mark as processing
		notification.Status = string(StatusProcessing)
		notification.UpdatedAt = time.Now()
		if err := s.scheduledNotifRepo.Update(ctx, notification); err != nil {
			s.log.Error("Failed to update notification status", logger.Fields{
				"notification_id": notification.ID.String(),
				"error":           err.Error(),
			})
			continue
		}

		// Process the notification
		err := s.processNotification(ctx, notification)
		now := time.Now()

		// Update status based on result
		if err != nil {
			notification.Status = string(StatusFailed)
			notification.Error = err.Error()
		} else {
			notification.Status = string(StatusSent)
			notification.SentAt = &now
		}

		notification.UpdatedAt = now
		if err := s.scheduledNotifRepo.Update(ctx, notification); err != nil {
			s.log.Error("Failed to update notification after processing", logger.Fields{
				"notification_id": notification.ID.String(),
				"error":           err.Error(),
			})
		}
	}
}

// processNotification sends a single scheduled notification
func (s *Service) processNotification(ctx context.Context, notification *model.ScheduledNotification) error {
	switch NotificationType(notification.Type) {
	case TypeSMS:
		return s.processSMSNotification(ctx, notification)
	case TypePush:
		return s.processPushNotification(ctx, notification)
	default:
		return errorx.New(errorx.BadRequest, "unsupported notification type: "+notification.Type)
	}
}

// processSMSNotification sends a scheduled SMS notification
func (s *Service) processSMSNotification(ctx context.Context, notification *model.ScheduledNotification) error {
	// Parse the SMS content
	var payload sms.MessagePayload
	if err := json.Unmarshal([]byte(notification.Content), &payload); err != nil {
		return errorx.Wrap(err, "failed to unmarshal SMS content")
	}

	provider := sms.SMSProvider(notification.Provider)
	if provider == "" {
		provider = sms.ProviderTwilio
	}

	// Send to user or phone number directly
	if notification.UserID != uuid.Nil {
		_, err := s.smsService.SendToUser(ctx, notification.UserID, &payload, provider)
		return err
	} else if notification.Recipient != "" {
		_, err := s.smsService.SendToNumber(ctx, notification.Recipient, &payload, provider)
		return err
	}

	return errorx.New(errorx.BadRequest, "notification has neither user ID nor recipient")
}

// processPushNotification sends a scheduled push notification
func (s *Service) processPushNotification(ctx context.Context, notification *model.ScheduledNotification) error {
	// Parse the push content
	var payload push.NotificationPayload
	if err := json.Unmarshal([]byte(notification.Content), &payload); err != nil {
		return errorx.Wrap(err, "failed to unmarshal push content")
	}

	provider := push.PushProvider(notification.Provider)
	if provider == "" {
		provider = push.ProviderExpo
	}

	// Send to user
	if notification.UserID != uuid.Nil {
		_, err := s.pushService.SendToUser(ctx, notification.UserID, &payload, provider)
		return err
	}

	return errorx.New(errorx.BadRequest, "push notification requires a user ID")
}

// convertToScheduledNotification converts a model to a response format
func (s *Service) convertToScheduledNotification(
	notification *model.ScheduledNotification,
) (*ScheduledNotification, error) {
	var content interface{}
	var metadata map[string]interface{}

	switch NotificationType(notification.Type) {
	case TypeSMS:
		var payload sms.MessagePayload
		if err := json.Unmarshal([]byte(notification.Content), &payload); err != nil {
			return nil, errorx.Wrap(err, "failed to unmarshal SMS content")
		}
		content = payload
		if payload.Data != nil {
			metadata = map[string]interface{}{
				"data": payload.Data,
			}
		}
	case TypePush:
		var payload push.NotificationPayload
		if err := json.Unmarshal([]byte(notification.Content), &payload); err != nil {
			return nil, errorx.Wrap(err, "failed to unmarshal push content")
		}
		content = payload
		if payload.Data != nil {
			metadata = map[string]interface{}{
				"data": payload.Data,
			}
		}
	default:
		return nil, errorx.New(errorx.BadRequest, "unsupported notification type: "+notification.Type)
	}

	return &ScheduledNotification{
		ID:         notification.ID,
		UserID:     notification.UserID,
		Recipient:  notification.Recipient,
		Type:       NotificationType(notification.Type),
		Status:     NotificationStatus(notification.Status),
		Provider:   notification.Provider,
		Content:    content,
		Metadata:   metadata,
		ScheduleAt: notification.ScheduleAt,
		SentAt:     notification.SentAt,
		CreatedAt:  notification.CreatedAt,
		UpdatedAt:  notification.UpdatedAt,
		Error:      notification.Error,
	}, nil
}
