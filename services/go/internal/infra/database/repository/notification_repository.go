package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mamacare/services/internal/domain/model"
	"github.com/mamacare/services/internal/domain/repository"
	"github.com/mamacare/services/internal/infra/database"
	"github.com/mamacare/services/pkg/errorx"
	"github.com/mamacare/services/pkg/logger"
)

// NotificationRepository implements repository.NotificationRepository interface
type NotificationRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(pool *pgxpool.Pool, logger logger.Logger) repository.NotificationRepository {
	return &NotificationRepository{
		pool:   pool,
		logger: logger,
	}
}

// scanNotification scans a notification from a row
func scanNotification(row pgx.Row) (*model.Notification, error) {
	var notification model.Notification
	var readAt *time.Time
	var relatedEntityID *uuid.UUID

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Message,
		&notification.Type,
		&notification.Priority,
		&notification.IsRead,
		&readAt,
		&relatedEntityID,
		&notification.RelatedEntityType,
		&notification.CreatedAt,
		&notification.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errorx.New(errorx.NotFound, "notification not found")
		}
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan notification")
	}

	// Set optional fields
	notification.ReadAt = readAt
	notification.RelatedEntityID = relatedEntityID

	return &notification, nil
}

// FindByID retrieves a notification by ID
func (r *NotificationRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Notification, error) {
	query := `
		SELECT 
			n.id, 
			n.user_id, 
			n.title, 
			n.message, 
			n.type, 
			n.priority, 
			n.is_read, 
			n.read_at, 
			n.related_entity_id, 
			n.related_entity_type, 
			n.created_at, 
			n.updated_at
		FROM notifications n
		WHERE n.id = $1
	`

	row := database.GetQuerier(ctx, r.pool).QueryRow(ctx, query, id)
	return scanNotification(row)
}

// FindByUser retrieves notifications for a user
func (r *NotificationRepository) FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.Notification, error) {
	query := `
		SELECT 
			n.id, 
			n.user_id, 
			n.title, 
			n.message, 
			n.type, 
			n.priority, 
			n.is_read, 
			n.read_at, 
			n.related_entity_id, 
			n.related_entity_type, 
			n.created_at, 
			n.updated_at
		FROM notifications n
		WHERE n.user_id = $1
		ORDER BY n.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, userID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query notifications by user")
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// FindUnreadByUser retrieves unread notifications for a user
func (r *NotificationRepository) FindUnreadByUser(ctx context.Context, userID uuid.UUID) ([]*model.Notification, error) {
	query := `
		SELECT 
			n.id, 
			n.user_id, 
			n.title, 
			n.message, 
			n.type, 
			n.priority, 
			n.is_read, 
			n.read_at, 
			n.related_entity_id, 
			n.related_entity_type, 
			n.created_at, 
			n.updated_at
		FROM notifications n
		WHERE n.user_id = $1 AND n.is_read = false
		ORDER BY n.priority DESC, n.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, userID)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query unread notifications")
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// FindByType retrieves notifications by type
func (r *NotificationRepository) FindByType(ctx context.Context, notificationType model.NotificationType) ([]*model.Notification, error) {
	query := `
		SELECT 
			n.id, 
			n.user_id, 
			n.title, 
			n.message, 
			n.type, 
			n.priority, 
			n.is_read, 
			n.read_at, 
			n.related_entity_id, 
			n.related_entity_type, 
			n.created_at, 
			n.updated_at
		FROM notifications n
		WHERE n.type = $1
		ORDER BY n.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, notificationType)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query notifications by type")
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// FindByPriority retrieves notifications by priority
func (r *NotificationRepository) FindByPriority(ctx context.Context, priority model.NotificationPriority) ([]*model.Notification, error) {
	query := `
		SELECT 
			n.id, 
			n.user_id, 
			n.title, 
			n.message, 
			n.type, 
			n.priority, 
			n.is_read, 
			n.read_at, 
			n.related_entity_id, 
			n.related_entity_type, 
			n.created_at, 
			n.updated_at
		FROM notifications n
		WHERE n.priority = $1
		ORDER BY n.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, priority)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query notifications by priority")
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// FindByDateRange retrieves notifications created in a date range for a user
func (r *NotificationRepository) FindByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*model.Notification, error) {
	query := `
		SELECT 
			n.id, 
			n.user_id, 
			n.title, 
			n.message, 
			n.type, 
			n.priority, 
			n.is_read, 
			n.read_at, 
			n.related_entity_id, 
			n.related_entity_type, 
			n.created_at, 
			n.updated_at
		FROM notifications n
		WHERE n.user_id = $1 AND n.created_at BETWEEN $2 AND $3
		ORDER BY n.created_at DESC
	`

	rows, err := database.GetQuerier(ctx, r.pool).Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to query notifications by date range")
	}
	defer rows.Close()

	return scanNotifications(rows)
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE notifications
		SET is_read = true, read_at = $1, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()
	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, now, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to mark notification as read")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "notification not found")
	}

	return nil
}

// Save creates or updates a notification
func (r *NotificationRepository) Save(ctx context.Context, notification *model.Notification) error {
	// Update timestamp for modifications
	notification.UpdatedAt = time.Now()

	// Use upsert to handle both insert and update
	query := `
		INSERT INTO notifications (
			id, user_id, title, message, type, priority, is_read, read_at,
			related_entity_id, related_entity_type, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			title = EXCLUDED.title,
			message = EXCLUDED.message,
			type = EXCLUDED.type,
			priority = EXCLUDED.priority,
			is_read = EXCLUDED.is_read,
			read_at = EXCLUDED.read_at,
			related_entity_id = EXCLUDED.related_entity_id,
			related_entity_type = EXCLUDED.related_entity_type,
			updated_at = EXCLUDED.updated_at
	`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Title,
		notification.Message,
		notification.Type,
		notification.Priority,
		notification.IsRead,
		notification.ReadAt,
		notification.RelatedEntityID,
		notification.RelatedEntityType,
		notification.CreatedAt,
		notification.UpdatedAt,
	)

	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to save notification")
	}

	return nil
}

// Delete removes a notification
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`

	cmdTag, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, id)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete notification")
	}

	if cmdTag.RowsAffected() == 0 {
		return errorx.New(errorx.NotFound, "notification not found")
	}

	return nil
}

// DeleteAllByUser removes all notifications for a user
func (r *NotificationRepository) DeleteAllByUser(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM notifications WHERE user_id = $1`

	_, err := database.GetQuerier(ctx, r.pool).Exec(ctx, query, userID)
	if err != nil {
		return errorx.Wrap(err, errorx.InternalServerError, "failed to delete all notifications for user")
	}

	return nil
}

// scanNotifications scans multiple notifications from rows
func scanNotifications(rows pgx.Rows) ([]*model.Notification, error) {
	var notifications []*model.Notification

	for rows.Next() {
		var notification model.Notification
		var readAt *time.Time
		var relatedEntityID *uuid.UUID

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.Priority,
			&notification.IsRead,
			&readAt,
			&relatedEntityID,
			&notification.RelatedEntityType,
			&notification.CreatedAt,
			&notification.UpdatedAt,
		)

		if err != nil {
			return nil, errorx.Wrap(err, errorx.InternalServerError, "failed to scan notification")
		}

		// Set optional fields
		notification.ReadAt = readAt
		notification.RelatedEntityID = relatedEntityID

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		return nil, errorx.Wrap(err, errorx.InternalServerError, "error iterating over notification rows")
	}

	return notifications, nil
}
