package model

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// NotificationTypeAppointment for appointment reminders
	NotificationTypeAppointment NotificationType = "appointment"
	// NotificationTypeEmergency for emergency alerts
	NotificationTypeEmergency NotificationType = "emergency"
	// NotificationTypeGeneral for general announcements
	NotificationTypeGeneral NotificationType = "general"
	// NotificationTypeHealthTip for health tips and advice
	NotificationTypeHealthTip NotificationType = "health_tip"
	// NotificationTypeSystem for system notifications
	NotificationTypeSystem NotificationType = "system"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	// NotificationPriorityLow for low priority notifications
	NotificationPriorityLow NotificationPriority = "low"
	// NotificationPriorityMedium for medium priority notifications
	NotificationPriorityMedium NotificationPriority = "medium"
	// NotificationPriorityHigh for high priority notifications
	NotificationPriorityHigh NotificationPriority = "high"
	// NotificationPriorityCritical for critical notifications
	NotificationPriorityCritical NotificationPriority = "critical"
)

// EntityType represents the type of entity related to a notification
type EntityType string

const (
	// EntityTypeVisit for visit-related notifications
	EntityTypeVisit EntityType = "visit"
	// EntityTypeSOSEvent for SOS event-related notifications
	EntityTypeSOSEvent EntityType = "sos_event"
	// EntityTypeHealthMetric for health metric-related notifications
	EntityTypeHealthMetric EntityType = "health_metric"
	// EntityTypeEducationContent for education content-related notifications
	EntityTypeEducationContent EntityType = "education_content"
)

// Notification represents a user notification
type Notification struct {
	ID                uuid.UUID           `json:"id"`
	UserID            uuid.UUID           `json:"user_id"`
	Title             string              `json:"title"`
	Message           string              `json:"message"`
	Type              NotificationType    `json:"type"`
	Priority          NotificationPriority `json:"priority"`
	IsRead            bool                `json:"is_read"`
	ReadAt            *time.Time          `json:"read_at,omitempty"`
	RelatedEntityID   *uuid.UUID          `json:"related_entity_id,omitempty"`
	RelatedEntityType EntityType          `json:"related_entity_type,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

// NewNotification creates a new notification
func NewNotification(id, userID uuid.UUID, title, message string, notificationType NotificationType) *Notification {
	now := time.Now()
	return &Notification{
		ID:        id,
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      notificationType,
		Priority:  NotificationPriorityMedium, // Default priority
		IsRead:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// WithPriority sets the priority of the notification
func (n *Notification) WithPriority(priority NotificationPriority) *Notification {
	n.Priority = priority
	return n
}

// WithRelatedEntity adds related entity information to the notification
func (n *Notification) WithRelatedEntity(entityID uuid.UUID, entityType EntityType) *Notification {
	n.RelatedEntityID = &entityID
	n.RelatedEntityType = entityType
	return n
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	if !n.IsRead {
		n.IsRead = true
		now := time.Now()
		n.ReadAt = &now
		n.UpdatedAt = now
	}
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread() {
	if n.IsRead {
		n.IsRead = false
		n.ReadAt = nil
		n.UpdatedAt = time.Now()
	}
}

// IsPriority checks if the notification is high priority or critical
func (n *Notification) IsPriority() bool {
	return n.Priority == NotificationPriorityHigh || n.Priority == NotificationPriorityCritical
}

// IsEmergency checks if the notification is an emergency
func (n *Notification) IsEmergency() bool {
	return n.Type == NotificationTypeEmergency
}

// Age returns the time elapsed since the notification was created
func (n *Notification) Age() time.Duration {
	return time.Since(n.CreatedAt)
}

// IsRecent checks if the notification is recent (created within the specified duration)
func (n *Notification) IsRecent(duration time.Duration) bool {
	return n.Age() <= duration
}
