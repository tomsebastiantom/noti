package implementations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/internal/webhooks/domain"
	"getnoti.com/internal/webhooks/repos"
	"getnoti.com/pkg/db"
	"github.com/google/uuid"
)

type sqlWebhookRepository struct {
	db db.Database
}

// NewWebhookRepository creates a new SQL webhook repository
func NewWebhookRepository(database db.Database) repos.WebhookRepository {
	return &sqlWebhookRepository{db: database}
}

// CreateWebhook creates a new webhook in the database
func (r *sqlWebhookRepository) CreateWebhook(ctx context.Context, webhook *domain.Webhook) (*domain.Webhook, error) {
	if webhook.ID == "" {
		webhook.ID = uuid.New().String()
	}
	webhook.CreatedAt = time.Now()
	webhook.UpdatedAt = time.Now()

	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		INSERT INTO webhooks (id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(ctx, query,
		webhook.ID,
		webhook.TenantID,
		webhook.URL,
		webhook.Secret,
		string(eventsJSON),
		string(headersJSON),
		webhook.IsActive,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return webhook, nil
}

// GetWebhookByID retrieves a webhook by ID
func (r *sqlWebhookRepository) GetWebhookByID(ctx context.Context, webhookID string) (*domain.Webhook, error) {
	query := `
		SELECT id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at, last_used_at
		FROM webhooks
		WHERE id = ?
	`

	row := r.db.QueryRow(ctx, query, webhookID)
	return r.scanWebhook(row)
}

// GetWebhooksByTenantID retrieves all webhooks for a specific tenant
func (r *sqlWebhookRepository) GetWebhooksByTenantID(ctx context.Context, tenantID string) ([]*domain.Webhook, error) {
	query := `
		SELECT id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at, last_used_at
		FROM webhooks
		WHERE tenant_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks for tenant %s: %w", tenantID, err)
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		webhook, err := r.scanWebhook(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhook rows: %w", err)
	}

	return webhooks, nil
}

// ListWebhooks retrieves webhooks with pagination
func (r *sqlWebhookRepository) ListWebhooks(ctx context.Context, limit, offset int) ([]*domain.Webhook, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM webhooks`
	var total int64
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get webhook count: %w", err)
	}

	// Get webhooks with pagination
	query := `
		SELECT id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at, last_used_at
		FROM webhooks
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		webhook, err := r.scanWebhook(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating webhook rows: %w", err)
	}

	return webhooks, total, nil
}

// GetWebhooksByEventType retrieves webhooks that listen to a specific event type
func (r *sqlWebhookRepository) GetWebhooksByEventType(ctx context.Context, eventType string) ([]*domain.Webhook, error) {
	query := `
		SELECT id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at, last_used_at
		FROM webhooks
		WHERE is_active = true AND (events LIKE ? OR events LIKE ?)
	`

	rows, err := r.db.Query(ctx, query, "%"+eventType+"%", "%*%")
	if err != nil {
		return nil, fmt.Errorf("failed to query webhooks by event type: %w", err)
	}
	defer rows.Close()

	var webhooks []*domain.Webhook
	for rows.Next() {
		webhook, err := r.scanWebhook(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook: %w", err)
		}
		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhook rows: %w", err)
	}

	return webhooks, nil
}

// UpdateWebhook updates an existing webhook
func (r *sqlWebhookRepository) UpdateWebhook(ctx context.Context, webhook *domain.Webhook) (*domain.Webhook, error) {
	webhook.UpdatedAt = time.Now()

	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	headersJSON, err := json.Marshal(webhook.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		UPDATE webhooks 
		SET url = ?, secret = ?, events = ?, headers = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(ctx, query,
		webhook.URL,
		webhook.Secret,
		string(eventsJSON),
		string(headersJSON),
		webhook.IsActive,
		webhook.UpdatedAt,
		webhook.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	return webhook, nil
}

// DeleteWebhook deletes a webhook
func (r *sqlWebhookRepository) DeleteWebhook(ctx context.Context, webhookID string) error {
	query := `DELETE FROM webhooks WHERE id = ?`
	_, err := r.db.Exec(ctx, query, webhookID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}

// UpdateWebhookLastUsed updates the last used timestamp
func (r *sqlWebhookRepository) UpdateWebhookLastUsed(ctx context.Context, webhookID string) error {
	now := time.Now()
	query := `UPDATE webhooks SET last_used_at = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, now, now, webhookID)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}

// SetWebhookActive sets the active status of a webhook
func (r *sqlWebhookRepository) SetWebhookActive(ctx context.Context, webhookID string, active bool) error {
	now := time.Now()
	query := `UPDATE webhooks SET is_active = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, active, now, webhookID)
	if err != nil {
		return fmt.Errorf("failed to set webhook active status: %w", err)
	}
	return nil
}

// Simplified implementations for delivery and event operations
func (r *sqlWebhookRepository) CreateDelivery(ctx context.Context, delivery *domain.WebhookDelivery) (*domain.WebhookDelivery, error) {
	// Implementation can be added later when needed
	return delivery, nil
}

func (r *sqlWebhookRepository) GetDeliveryByID(ctx context.Context, deliveryID string) (*domain.WebhookDelivery, error) {
	// Implementation can be added later when needed
	return nil, fmt.Errorf("not implemented")
}

func (r *sqlWebhookRepository) ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*domain.WebhookDelivery, int64, error) {
	// Implementation can be added later when needed
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *sqlWebhookRepository) GetPendingDeliveries(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error) {
	// Implementation can be added later when needed
	return nil, fmt.Errorf("not implemented")
}

func (r *sqlWebhookRepository) UpdateDelivery(ctx context.Context, delivery *domain.WebhookDelivery) (*domain.WebhookDelivery, error) {
	// Implementation can be added later when needed
	return delivery, nil
}

func (r *sqlWebhookRepository) CreateEvent(ctx context.Context, event *domain.WebhookEvent) (*domain.WebhookEvent, error) {
	// Implementation can be added later when needed
	return event, nil
}

func (r *sqlWebhookRepository) UpdateEvent(ctx context.Context, event *domain.WebhookEvent) (*domain.WebhookEvent, error) {
	// Implementation can be added later when needed
	return event, nil
}

func (r *sqlWebhookRepository) ListEvents(ctx context.Context, webhookID string, limit, offset int) ([]*domain.WebhookEvent, int64, error) {
	// Implementation can be added later when needed
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *sqlWebhookRepository) GetStats(ctx context.Context) (*domain.WebhookStats, error) {
	// Implementation can be added later when needed
	return &domain.WebhookStats{}, nil
}

// scanWebhook scans a webhook from a database row
func (r *sqlWebhookRepository) scanWebhook(scanner interface{}) (*domain.Webhook, error) {
	var webhook domain.Webhook
	var eventsJSON, headersJSON string
	var lastUsedAt sql.NullTime

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&webhook.ID,
			&webhook.TenantID,
			&webhook.URL,
			&webhook.Secret,
			&eventsJSON,
			&headersJSON,
			&webhook.IsActive,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
			&lastUsedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&webhook.ID,
			&webhook.TenantID,
			&webhook.URL,
			&webhook.Secret,
			&eventsJSON,
			&headersJSON,
			&webhook.IsActive,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
			&lastUsedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan webhook: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(eventsJSON), &webhook.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	if err := json.Unmarshal([]byte(headersJSON), &webhook.Headers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal headers: %w", err)
	}

	// Handle nullable last used at
	if lastUsedAt.Valid {
		webhook.LastUsedAt = &lastUsedAt.Time
	}

	return &webhook, nil
}

// scanDelivery scans a webhook delivery from a database row
func (r *sqlWebhookRepository) scanDelivery(scanner interface{}) (*domain.WebhookDelivery, error) {
	var delivery domain.WebhookDelivery
	var deliveredAt, nextRetryAt sql.NullTime

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&delivery.ID,
			&delivery.WebhookID,
			&delivery.TenantID,
			&delivery.EventType,
			&delivery.EventID,
			&delivery.Payload,
			&delivery.StatusCode,
			&delivery.Response,
			&delivery.AttemptCount,
			&deliveredAt,
			&delivery.CreatedAt,
			&nextRetryAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&delivery.ID,
			&delivery.WebhookID,
			&delivery.TenantID,
			&delivery.EventType,
			&delivery.EventID,
			&delivery.Payload,
			&delivery.StatusCode,
			&delivery.Response,
			&delivery.AttemptCount,
			&deliveredAt,
			&delivery.CreatedAt,
			&nextRetryAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan delivery: %w", err)
	}

	// Handle nullable fields
	if deliveredAt.Valid {
		delivery.DeliveredAt = &deliveredAt.Time
	}
	if nextRetryAt.Valid {
		delivery.NextRetryAt = &nextRetryAt.Time
	}

	return &delivery, nil
}

// scanEvent scans a webhook event from a database row
func (r *sqlWebhookRepository) scanEvent(scanner interface{}) (*domain.WebhookEvent, error) {
	var event domain.WebhookEvent
	var eventDataJSON string
	var deliveryID sql.NullString

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(
			&event.ID,
			&event.WebhookID,
			&event.TenantID,
			&event.EventType,
			&eventDataJSON,
			&deliveryID,
			&event.CreatedAt,
		)
	case *sql.Rows:
		err = s.Scan(
			&event.ID,
			&event.WebhookID,
			&event.TenantID,
			&event.EventType,
			&eventDataJSON,
			&deliveryID,
			&event.CreatedAt,
		)
	default:
		return nil, fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	// Parse JSON event data
	if err := json.Unmarshal([]byte(eventDataJSON), &event.EventData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	// Handle nullable delivery ID
	if deliveryID.Valid {
		event.DeliveryID = &deliveryID.String
	}

	return &event, nil
}
