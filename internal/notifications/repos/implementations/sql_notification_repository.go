package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"getnoti.com/internal/notifications/domain"
	"getnoti.com/internal/notifications/repos"
	"getnoti.com/pkg/db"
)

type sqlNotificationRepository struct {
	db db.Database
}

// NewNotificationRepository creates a new instance of sqlNotificationRepository
func NewNotificationRepository(db db.Database) repository.NotificationRepository {
	return &sqlNotificationRepository{db: db}
}

// CreateNotification inserts a new notification into the database
func (r *sqlNotificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	variables, err := json.Marshal(notification.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `INSERT INTO notifications (id, tenant_id, user_id, type, channel, template_id, status, content, variables) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.Exec(ctx, query, notification.ID, notification.TenantID, notification.UserID, notification.Type, notification.Channel, notification.TemplateID, notification.Status, notification.Content, variables)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetNotificationByID retrieves a notification by its ID
func (r *sqlNotificationRepository) GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, type, channel, template_id, status, content, variables FROM notifications WHERE id = ?`
	row := r.db.QueryRow(ctx, query, id)
	notification := &domain.Notification{}
	var variables []byte
	err := row.Scan(&notification.ID, &notification.TenantID, &notification.UserID, &notification.Type, &notification.Channel, &notification.TemplateID, &notification.Status, &notification.Content, &variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	err = json.Unmarshal(variables, &notification.Variables)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	return notification, nil
}

// UpdateNotification updates an existing notification in the database
func (r *sqlNotificationRepository) UpdateNotification(ctx context.Context, notification *domain.Notification) error {
	variables, err := json.Marshal(notification.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	query := `UPDATE notifications SET tenant_id = ?, user_id = ?, type = ?, channel = ?, template_id = ?, status = ?, content = ?, variables = ? WHERE id = ?`
	_, err = r.db.Exec(ctx, query, notification.TenantID, notification.UserID, notification.Type, notification.Channel, notification.TemplateID, notification.Status, notification.Content, variables, notification.ID)
	if err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}
	return nil
}

// DeleteNotification deletes a notification from the database
func (r *sqlNotificationRepository) DeleteNotification(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = ?`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}
