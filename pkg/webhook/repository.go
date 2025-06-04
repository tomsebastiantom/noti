package webhook

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/pkg/db"
	"getnoti.com/pkg/errors"
	"getnoti.com/pkg/logger"
)

// Repository handles webhook persistence operations
type Repository interface {
	// Webhook configuration management
	CreateWebhook(ctx context.Context, webhook *WebhookConfig) error
	GetWebhook(ctx context.Context, tenantID, webhookID string) (*WebhookConfig, error)
	GetWebhooksByTenant(ctx context.Context, tenantID string) ([]*WebhookConfig, error)
	UpdateWebhook(ctx context.Context, webhook *WebhookConfig) error
	DeleteWebhook(ctx context.Context, tenantID, webhookID string) error
	
	// Webhook delivery management
	CreateDelivery(ctx context.Context, delivery *WebhookDelivery) error
	GetDelivery(ctx context.Context, tenantID, deliveryID string) (*WebhookDelivery, error)
	GetDeliveriesByWebhook(ctx context.Context, tenantID, webhookID string, limit int) ([]*WebhookDelivery, error)
	UpdateDelivery(ctx context.Context, delivery *WebhookDelivery) error
	GetPendingRetries(ctx context.Context, tenantID string, limit int) ([]*WebhookDelivery, error)
}

// SQLRepository implements Repository using SQL database
type SQLRepository struct {
	db     db.Database
	logger logger.Logger
}

// NewSQLRepository creates a new SQL webhook repository
func NewSQLRepository(database db.Database, logger logger.Logger) Repository {
	return &SQLRepository{
		db:     database,
		logger: logger,
	}
}

// StringSlice is a custom type for handling string slices in database
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(s)
	return string(data), err
}

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}
	
	return json.Unmarshal(bytes, s)
}

// HeaderMap is a custom type for handling map[string]string in database
type HeaderMap map[string]string

func (h HeaderMap) Value() (driver.Value, error) {
	if len(h) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(h)
	return string(data), err
}

func (h *HeaderMap) Scan(value interface{}) error {
	if value == nil {
		*h = make(map[string]string)
		return nil
	}
	
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into HeaderMap", value)
	}
	
	return json.Unmarshal(bytes, h)
}

// CreateWebhook creates a new webhook configuration
func (r *SQLRepository) CreateWebhook(ctx context.Context, webhook *WebhookConfig) error {
	r.logger.DebugContext(ctx, "Creating webhook",
		logger.String("tenant_id", webhook.TenantID),
		logger.String("webhook_id", webhook.ID))

	query := `
		INSERT INTO webhooks (id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	events := StringSlice(webhook.Events)
	headers := HeaderMap(webhook.Headers)
	
	_, err := r.db.Exec(ctx, query,
		webhook.ID,
		webhook.TenantID,
		webhook.URL,
		webhook.Secret,
		events,
		headers,
		webhook.IsActive,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	)

	if err != nil {
		return errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("CreateWebhook").
			WithCause(err).
			WithMessage("Failed to create webhook").
			WithDetails(map[string]interface{}{
				"tenant_id":  webhook.TenantID,
				"webhook_id": webhook.ID,
			}).
			Build()
	}

	r.logger.InfoContext(ctx, "Webhook created successfully",
		logger.String("tenant_id", webhook.TenantID),
		logger.String("webhook_id", webhook.ID))

	return nil
}

// GetWebhook retrieves a webhook by tenant and webhook ID
func (r *SQLRepository) GetWebhook(ctx context.Context, tenantID, webhookID string) (*WebhookConfig, error) {
	r.logger.DebugContext(ctx, "Getting webhook",
		logger.String("tenant_id", tenantID),
		logger.String("webhook_id", webhookID))

	query := `
		SELECT id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at, last_used_at
		FROM webhooks 
		WHERE tenant_id = ? AND id = ?`

	webhook := &WebhookConfig{}
	var events StringSlice
	var headers HeaderMap
	var lastUsedAt *time.Time

	err := r.db.QueryRow(ctx, query, tenantID, webhookID).Scan(
		&webhook.ID,
		&webhook.TenantID,
		&webhook.URL,
		&webhook.Secret,
		&events,
		&headers,
		&webhook.IsActive,
		&webhook.CreatedAt,
		&webhook.UpdatedAt,
		&lastUsedAt,	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ErrCodeNotFound).
				WithContext(ctx).
				WithOperation("GetWebhook").
				WithMessage("Webhook not found").
				WithDetails(map[string]interface{}{
					"tenant_id":  tenantID,
					"webhook_id": webhookID,
				}).
				Build()
		}
		return nil, errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("GetWebhook").
			WithCause(err).
			WithMessage("Failed to get webhook").
			WithDetails(map[string]interface{}{
				"tenant_id":  tenantID,
				"webhook_id": webhookID,
			}).
			Build()
	}

	webhook.Events = []string(events)
	webhook.Headers = map[string]string(headers)
	webhook.LastUsedAt = lastUsedAt

	return webhook, nil
}

// GetWebhooksByTenant retrieves all webhooks for a tenant
func (r *SQLRepository) GetWebhooksByTenant(ctx context.Context, tenantID string) ([]*WebhookConfig, error) {
	r.logger.DebugContext(ctx, "Getting webhooks for tenant",
		logger.String("tenant_id", tenantID))

	query := `
		SELECT id, tenant_id, url, secret, events, headers, is_active, created_at, updated_at, last_used_at
		FROM webhooks 
		WHERE tenant_id = ? 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("GetWebhooksByTenant").
			WithCause(err).
			WithMessage("Failed to get webhooks").
			WithDetails(map[string]interface{}{
				"tenant_id": tenantID,
			}).
			Build()
	}
	defer rows.Close()

	var webhooks []*WebhookConfig
	for rows.Next() {
		webhook := &WebhookConfig{}
		var events StringSlice
		var headers HeaderMap
		var lastUsedAt *time.Time

		err := rows.Scan(
			&webhook.ID,
			&webhook.TenantID,
			&webhook.URL,
			&webhook.Secret,
			&events,
			&headers,
			&webhook.IsActive,
			&webhook.CreatedAt,
			&webhook.UpdatedAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, errors.New(errors.ErrCodeDatabase).
				WithContext(ctx).
				WithOperation("GetWebhooksByTenant").
				WithCause(err).
				WithMessage("Failed to scan webhook").
				WithDetails(map[string]interface{}{
					"tenant_id": tenantID,
				}).
				Build()
		}

		webhook.Events = []string(events)
		webhook.Headers = map[string]string(headers)
		webhook.LastUsedAt = lastUsedAt
		webhooks = append(webhooks, webhook)
	}

	return webhooks, nil
}

// UpdateWebhook updates an existing webhook
func (r *SQLRepository) UpdateWebhook(ctx context.Context, webhook *WebhookConfig) error {
	r.logger.DebugContext(ctx, "Updating webhook",
		logger.String("tenant_id", webhook.TenantID),
		logger.String("webhook_id", webhook.ID))

	query := `
		UPDATE webhooks 
		SET url = ?, events = ?, headers = ?, is_active = ?, updated_at = ?
		WHERE tenant_id = ? AND id = ?`

	events := StringSlice(webhook.Events)
	headers := HeaderMap(webhook.Headers)
	webhook.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		webhook.URL,
		events,
		headers,
		webhook.IsActive,
		webhook.UpdatedAt,
		webhook.TenantID,
		webhook.ID,
	)

	if err != nil {
		return errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("UpdateWebhook").
			WithCause(err).
			WithMessage("Failed to update webhook").
			WithDetails(map[string]interface{}{
				"tenant_id":  webhook.TenantID,
				"webhook_id": webhook.ID,
			}).
			Build()
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ErrCodeNotFound).
			WithContext(ctx).
			WithOperation("UpdateWebhook").
			WithMessage("Webhook not found for update").
			WithDetails(map[string]interface{}{
				"tenant_id":  webhook.TenantID,
				"webhook_id": webhook.ID,
			}).
			Build()
	}

	r.logger.InfoContext(ctx, "Webhook updated successfully",
		logger.String("tenant_id", webhook.TenantID),
		logger.String("webhook_id", webhook.ID))

	return nil
}

// DeleteWebhook deletes a webhook
func (r *SQLRepository) DeleteWebhook(ctx context.Context, tenantID, webhookID string) error {
	r.logger.DebugContext(ctx, "Deleting webhook",
		logger.String("tenant_id", tenantID),
		logger.String("webhook_id", webhookID))

	query := `DELETE FROM webhooks WHERE tenant_id = ? AND id = ?`

	result, err := r.db.Exec(ctx, query, tenantID, webhookID)
	if err != nil {
		return errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("DeleteWebhook").
			WithCause(err).
			WithMessage("Failed to delete webhook").
			WithDetails(map[string]interface{}{
				"tenant_id":  tenantID,
				"webhook_id": webhookID,
			}).
			Build()
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ErrCodeNotFound).
			WithContext(ctx).
			WithOperation("DeleteWebhook").
			WithMessage("Webhook not found for deletion").
			WithDetails(map[string]interface{}{
				"tenant_id":  tenantID,
				"webhook_id": webhookID,
			}).
			Build()
	}

	r.logger.InfoContext(ctx, "Webhook deleted successfully",
		logger.String("tenant_id", tenantID),
		logger.String("webhook_id", webhookID))

	return nil
}

// CreateDelivery creates a new webhook delivery record
func (r *SQLRepository) CreateDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	r.logger.DebugContext(ctx, "Creating webhook delivery",
		logger.String("tenant_id", delivery.TenantID),
		logger.String("webhook_id", delivery.WebhookID),
		logger.String("delivery_id", delivery.ID))

	query := `
		INSERT INTO webhook_deliveries 
		(id, webhook_id, tenant_id, event_type, event_id, payload, status_code, response, attempt_count, delivered_at, created_at, next_retry_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(ctx, query,
		delivery.ID,
		delivery.WebhookID,
		delivery.TenantID,
		delivery.EventType,
		delivery.EventID,
		delivery.Payload,
		delivery.StatusCode,
		delivery.Response,
		delivery.AttemptCount,
		delivery.DeliveredAt,
		delivery.CreatedAt,
		delivery.NextRetryAt,
	)

	if err != nil {
		return errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("CreateDelivery").
			WithCause(err).
			WithMessage("Failed to create webhook delivery").
			WithDetails(map[string]interface{}{
				"tenant_id":   delivery.TenantID,
				"webhook_id":  delivery.WebhookID,
				"delivery_id": delivery.ID,
			}).
			Build()
	}

	return nil
}

// GetDelivery retrieves a webhook delivery by ID
func (r *SQLRepository) GetDelivery(ctx context.Context, tenantID, deliveryID string) (*WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, tenant_id, event_type, event_id, payload, status_code, response, attempt_count, delivered_at, created_at, next_retry_at
		FROM webhook_deliveries 
		WHERE tenant_id = ? AND id = ?`

	delivery := &WebhookDelivery{}
	err := r.db.QueryRow(ctx, query, tenantID, deliveryID).Scan(
		&delivery.ID,
		&delivery.WebhookID,
		&delivery.TenantID,
		&delivery.EventType,
		&delivery.EventID,
		&delivery.Payload,
		&delivery.StatusCode,
		&delivery.Response,
		&delivery.AttemptCount,
		&delivery.DeliveredAt,
		&delivery.CreatedAt,
		&delivery.NextRetryAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ErrCodeNotFound).
				WithContext(ctx).
				WithOperation("GetDelivery").
				WithMessage("Webhook delivery not found").
				WithDetails(map[string]interface{}{
					"tenant_id":   tenantID,
					"delivery_id": deliveryID,
				}).
				Build()
		}
		return nil, errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("GetDelivery").
			WithCause(err).
			WithMessage("Failed to get webhook delivery").
			WithDetails(map[string]interface{}{
				"tenant_id":   tenantID,
				"delivery_id": deliveryID,
			}).
			Build()
	}

	return delivery, nil
}

// GetDeliveriesByWebhook retrieves webhook deliveries for a specific webhook
func (r *SQLRepository) GetDeliveriesByWebhook(ctx context.Context, tenantID, webhookID string, limit int) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, tenant_id, event_type, event_id, payload, status_code, response, attempt_count, delivered_at, created_at, next_retry_at
		FROM webhook_deliveries 
		WHERE tenant_id = ? AND webhook_id = ?
		ORDER BY created_at DESC
		LIMIT ?`

	rows, err := r.db.Query(ctx, query, tenantID, webhookID, limit)
	if err != nil {
		return nil, errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("GetDeliveriesByWebhook").
			WithCause(err).
			WithMessage("Failed to get webhook deliveries").
			WithDetails(map[string]interface{}{
				"tenant_id":  tenantID,
				"webhook_id": webhookID,
			}).
			Build()
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		delivery := &WebhookDelivery{}
		err := rows.Scan(
			&delivery.ID,
			&delivery.WebhookID,
			&delivery.TenantID,
			&delivery.EventType,
			&delivery.EventID,
			&delivery.Payload,
			&delivery.StatusCode,
			&delivery.Response,
			&delivery.AttemptCount,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
			&delivery.NextRetryAt,
		)
		if err != nil {
			return nil, errors.New(errors.ErrCodeDatabase).
				WithContext(ctx).
				WithOperation("GetDeliveriesByWebhook").
				WithCause(err).
				WithMessage("Failed to scan webhook delivery").
				WithDetails(map[string]interface{}{
					"tenant_id":  tenantID,
					"webhook_id": webhookID,
				}).
				Build()
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, nil
}

// UpdateDelivery updates a webhook delivery record
func (r *SQLRepository) UpdateDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	query := `
		UPDATE webhook_deliveries 
		SET status_code = ?, response = ?, attempt_count = ?, delivered_at = ?, next_retry_at = ?
		WHERE tenant_id = ? AND id = ?`

	result, err := r.db.Exec(ctx, query,
		delivery.StatusCode,
		delivery.Response,
		delivery.AttemptCount,
		delivery.DeliveredAt,
		delivery.NextRetryAt,
		delivery.TenantID,
		delivery.ID,
	)

	if err != nil {
		return errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("UpdateDelivery").
			WithCause(err).
			WithMessage("Failed to update webhook delivery").
			WithDetails(map[string]interface{}{
				"tenant_id":   delivery.TenantID,
				"delivery_id": delivery.ID,
			}).
			Build()
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ErrCodeNotFound).
			WithContext(ctx).
			WithOperation("UpdateDelivery").
			WithMessage("Webhook delivery not found for update").
			WithDetails(map[string]interface{}{
				"tenant_id":   delivery.TenantID,
				"delivery_id": delivery.ID,
			}).
			Build()
	}

	return nil
}

// GetPendingRetries retrieves webhook deliveries that need to be retried
func (r *SQLRepository) GetPendingRetries(ctx context.Context, tenantID string, limit int) ([]*WebhookDelivery, error) {
	query := `
		SELECT id, webhook_id, tenant_id, event_type, event_id, payload, status_code, response, attempt_count, delivered_at, created_at, next_retry_at
		FROM webhook_deliveries 
		WHERE tenant_id = ? AND delivered_at IS NULL AND next_retry_at <= ?
		ORDER BY next_retry_at ASC
		LIMIT ?`

	rows, err := r.db.Query(ctx, query, tenantID, time.Now(), limit)
	if err != nil {
		return nil, errors.New(errors.ErrCodeDatabase).
			WithContext(ctx).
			WithOperation("GetPendingRetries").
			WithCause(err).
			WithMessage("Failed to get pending webhook retries").
			WithDetails(map[string]interface{}{
				"tenant_id": tenantID,
			}).
			Build()
	}
	defer rows.Close()

	var deliveries []*WebhookDelivery
	for rows.Next() {
		delivery := &WebhookDelivery{}
		err := rows.Scan(
			&delivery.ID,
			&delivery.WebhookID,
			&delivery.TenantID,
			&delivery.EventType,
			&delivery.EventID,
			&delivery.Payload,
			&delivery.StatusCode,
			&delivery.Response,
			&delivery.AttemptCount,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
			&delivery.NextRetryAt,
		)
		if err != nil {
			return nil, errors.New(errors.ErrCodeDatabase).
				WithContext(ctx).
				WithOperation("GetPendingRetries").
				WithCause(err).
				WithMessage("Failed to scan webhook delivery").
				WithDetails(map[string]interface{}{
					"tenant_id": tenantID,
				}).
				Build()
		}
		deliveries = append(deliveries, delivery)
	}

	return deliveries, nil
}
