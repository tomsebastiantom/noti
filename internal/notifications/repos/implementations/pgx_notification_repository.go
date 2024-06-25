package postgres

import (
	"context"
	"encoding/json"

	"getnoti.com/internal/notifications/domain"
	"getnoti.com/internal/notifications/repos"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresNotificationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresNotificationRepository(db *pgxpool.Pool) repository.NotificationRepository {
	return &postgresNotificationRepository{db: db}
}

func (r *postgresNotificationRepository) CreateNotification(ctx context.Context, notification *domain.Notification) error {
	variables, err := json.Marshal(notification.Variables)
	if err != nil {
		return err
	}

	query := `INSERT INTO notifications (id, tenant_id, user_id, type, channel, template_id, status, content, variables) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = r.db.Exec(ctx, query, notification.ID, notification.TenantID, notification.UserID, notification.Type, notification.Channel, notification.TemplateID, notification.Status, notification.Content, variables)
	return err
}

func (r *postgresNotificationRepository) GetNotificationByID(ctx context.Context, id string) (*domain.Notification, error) {
	query := `SELECT id, tenant_id, user_id, type, channel, template_id, status, content, variables FROM notifications WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	notification := &domain.Notification{}
	var variables []byte
	err := row.Scan(&notification.ID, &notification.TenantID, &notification.UserID, &notification.Type, &notification.Channel, &notification.TemplateID, &notification.Status, &notification.Content, &variables)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(variables, &notification.Variables)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (r *postgresNotificationRepository) UpdateNotification(ctx context.Context, notification *domain.Notification) error {
	variables, err := json.Marshal(notification.Variables)
	if err != nil {
		return err
	}

	query := `UPDATE notifications SET tenant_id = $1, user_id = $2, type = $3, channel = $4, template_id = $5, status = $6, content = $7, variables = $8 WHERE id = $9`
	_, err = r.db.Exec(ctx, query, notification.TenantID, notification.UserID, notification.Type, notification.Channel, notification.TemplateID, notification.Status, notification.Content, variables, notification.ID)
	return err
}

func (r *postgresNotificationRepository) DeleteNotification(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
