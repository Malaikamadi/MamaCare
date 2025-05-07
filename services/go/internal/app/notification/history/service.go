package history

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// NotificationHistory represents a notification that has been sent
type NotificationHistory struct {
	ID         uuid.UUID              `json:"id"`
	UserID     uuid.UUID              `json:"user_id,omitempty"`
	RecipientID uuid.UUID             `json:"recipient_id,omitempty"`
	Recipient  string                 `json:"recipient,omitempty"` // Phone number, email, etc.
	Type       string                 `json:"type"`                // Push, SMS, email
	Category   string                 `json:"category,omitempty"`  // Health alert, visit reminder, etc.
	Channel    string                 `json:"channel"`             // Provider used (Twilio, Expo, etc.)
	Title      string                 `json:"title,omitempty"`
	Content    string                 `json:"content"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Status     string                 `json:"status"`
	Error      string                 `json:"error,omitempty"`
	SentAt     time.Time              `json:"sent_at"`
	DeliveredAt *time.Time            `json:"delivered_at,omitempty"`
	ReadAt     *time.Time             `json:"read_at,omitempty"`
	ResponseAt *time.Time             `json:"response_at,omitempty"`
	Response   string                 `json:"response,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// NotificationFilter is used to filter notifications
type NotificationFilter struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Type      *string    `json:"type,omitempty"`
	Category  *string    `json:"category,omitempty"`
	Channel   *string    `json:"channel,omitempty"`
	Status    *string    `json:"status,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// Service handles notification history
type Service struct {
	notifHistoryRepo repository.NotificationHistoryRepository
	log              logger.Logger
}

// NewService creates a new notification history service
func NewService(
	notifHistoryRepo repository.NotificationHistoryRepository,
	log logger.Logger,
) *Service {
	return &Service{
		notifHistoryRepo: notifHistoryRepo,
		log:              log,
	}
}

// RecordNotification records a new notification in the history
func (s *Service) RecordNotification(
	ctx context.Context,
	userID uuid.UUID,
	notificationType string,
	category string,
	channel string,
	title string,
	content string,
	metadata map[string]interface{},
	status string,
) (*NotificationHistory, error) {
	metadataStr := ""
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, errorx.Wrap(err, errorx.Internal, "failed to marshal metadata")
		}
		metadataStr = string(metadataBytes)
	}

	notification := &model.NotificationHistory{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      notificationType,
		Category:  category,
		Channel:   channel,
		Title:     title,
		Content:   content,
		Metadata:  metadataStr,
		Status:    status,
		SentAt:    time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.notifHistoryRepo.Create(ctx, notification); err != nil {
		return nil, errorx.Wrap(err, errorx.Internal, "failed to record notification")
	}

	var metadataMap map[string]interface{}
	if notification.Metadata != "" {
		if err := json.Unmarshal([]byte(notification.Metadata), &metadataMap); err != nil {
			s.log.Error("Failed to unmarshal metadata", logger.Fields{
				"notification_id": notification.ID.String(),
				"error":           err.Error(),
			})
		}
	}

	return &NotificationHistory{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Category:  notification.Category,
		Channel:   notification.Channel,
		Title:     notification.Title,
		Content:   notification.Content,
		Metadata:  metadataMap,
		Status:    notification.Status,
		Error:     notification.Error,
		SentAt:    notification.SentAt,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}, nil
}

// GetNotificationHistory retrieves notification history for a user
func (s *Service) GetNotificationHistory(
	ctx context.Context,
	filter *NotificationFilter,
) ([]*NotificationHistory, error) {
	// Apply default limit if not provided
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Convert filter to repository format
	repoFilter := &repository.NotificationHistoryFilter{
		UserID:    filter.UserID,
		Type:      filter.Type,
		Category:  filter.Category,
		Channel:   filter.Channel,
		Status:    filter.Status,
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Limit:     filter.Limit,
		Offset:    filter.Offset,
	}

	notifications, err := s.notifHistoryRepo.GetByFilter(ctx, repoFilter)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.Internal, "failed to get notification history")
	}

	result := make([]*NotificationHistory, 0, len(notifications))
	for _, notification := range notifications {
		var metadataMap map[string]interface{}
		if notification.Metadata != "" {
			if err := json.Unmarshal([]byte(notification.Metadata), &metadataMap); err != nil {
				s.log.Error("Failed to unmarshal metadata", logger.Fields{
					"notification_id": notification.ID.String(),
					"error":           err.Error(),
				})
			}
		}

		history := &NotificationHistory{
			ID:         notification.ID,
			UserID:     notification.UserID,
			RecipientID: notification.RecipientID,
			Recipient:  notification.Recipient,
			Type:       notification.Type,
			Category:   notification.Category,
			Channel:    notification.Channel,
			Title:      notification.Title,
			Content:    notification.Content,
			Metadata:   metadataMap,
			Status:     notification.Status,
			Error:      notification.Error,
			SentAt:     notification.SentAt,
			DeliveredAt: notification.DeliveredAt,
			ReadAt:     notification.ReadAt,
			ResponseAt: notification.ResponseAt,
			Response:   notification.Response,
			CreatedAt:  notification.CreatedAt,
			UpdatedAt:  notification.UpdatedAt,
		}

		result = append(result, history)
	}

	return result, nil
}

// MarkNotificationAsDelivered marks a notification as delivered
func (s *Service) MarkNotificationAsDelivered(
	ctx context.Context,
	notificationID uuid.UUID,
) error {
	notification, err := s.notifHistoryRepo.GetByID(ctx, notificationID)
	if err != nil {
		return errorx.Wrap(err, errorx.NotFound, "notification not found")
	}

	now := time.Now()
	notification.Status = "delivered"
	notification.DeliveredAt = &now
	notification.UpdatedAt = now

	if err := s.notifHistoryRepo.Update(ctx, notification); err != nil {
		return errorx.Wrap(err, errorx.Internal, "failed to update notification")
	}

	return nil
}

// MarkNotificationAsRead marks a notification as read
func (s *Service) MarkNotificationAsRead(
	ctx context.Context,
	notificationID uuid.UUID,
) error {
	notification, err := s.notifHistoryRepo.GetByID(ctx, notificationID)
	if err != nil {
		return errorx.Wrap(err, errorx.NotFound, "notification not found")
	}

	now := time.Now()
	notification.ReadAt = &now
	notification.UpdatedAt = now

	if err := s.notifHistoryRepo.Update(ctx, notification); err != nil {
		return errorx.Wrap(err, errorx.Internal, "failed to update notification")
	}

	return nil
}

// RecordNotificationResponse records a user's response to a notification
func (s *Service) RecordNotificationResponse(
	ctx context.Context,
	notificationID uuid.UUID,
	response string,
) error {
	notification, err := s.notifHistoryRepo.GetByID(ctx, notificationID)
	if err != nil {
		return errorx.Wrap(err, errorx.NotFound, "notification not found")
	}

	now := time.Now()
	notification.Response = response
	notification.ResponseAt = &now
	notification.UpdatedAt = now

	if err := s.notifHistoryRepo.Update(ctx, notification); err != nil {
		return errorx.Wrap(err, errorx.Internal, "failed to update notification")
	}

	return nil
}

// DeleteNotificationHistory deletes notification history for a user
func (s *Service) DeleteNotificationHistory(
	ctx context.Context,
	userID uuid.UUID,
	olderThan *time.Time,
) error {
	if olderThan == nil {
		// Default to deleting notifications older than 90 days
		days90 := time.Now().AddDate(0, 0, -90)
		olderThan = &days90
	}

	if err := s.notifHistoryRepo.DeleteByUserIDAndDate(ctx, userID, *olderThan); err != nil {
		return errorx.Wrap(err, errorx.Internal, "failed to delete notification history")
	}

	s.log.Info("Deleted notification history", logger.Fields{
		"user_id":     userID.String(),
		"older_than":  olderThan.Format(time.RFC3339),
	})

	return nil
}

// GetNotificationStats gets statistics about notifications
func (s *Service) GetNotificationStats(
	ctx context.Context,
	userID uuid.UUID,
	startDate time.Time,
	endDate time.Time,
) (map[string]int, error) {
	stats, err := s.notifHistoryRepo.GetStatsByUserID(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.Internal, "failed to get notification stats")
	}

	return stats, nil
}

// CountUnreadNotifications counts the number of unread notifications for a user
func (s *Service) CountUnreadNotifications(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	count, err := s.notifHistoryRepo.CountUnreadByUserID(ctx, userID)
	if err != nil {
		return 0, errorx.Wrap(err, errorx.Internal, "failed to count unread notifications")
	}

	return count, nil
}
