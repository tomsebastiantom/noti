package repos

import (
	"context"

	"getnoti.com/internal/webhooks/domain"
)

// WebhookRepository defines the interface for webhook persistence operations
type WebhookRepository interface {
	// Webhook CRUD operations
	CreateWebhook(ctx context.Context, webhook *domain.Webhook) (*domain.Webhook, error)
	GetWebhookByID(ctx context.Context, webhookID string) (*domain.Webhook, error)
	GetWebhooksByTenantID(ctx context.Context, tenantID string) ([]*domain.Webhook, error)
	ListWebhooks(ctx context.Context, limit, offset int) ([]*domain.Webhook, int64, error)
	GetWebhooksByEventType(ctx context.Context, eventType string) ([]*domain.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *domain.Webhook) (*domain.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID string) error
	
	// Webhook status operations
	UpdateWebhookLastUsed(ctx context.Context, webhookID string) error
	SetWebhookActive(ctx context.Context, webhookID string, active bool) error
	
	// Delivery operations
	CreateDelivery(ctx context.Context, delivery *domain.WebhookDelivery) (*domain.WebhookDelivery, error)
	GetDeliveryByID(ctx context.Context, deliveryID string) (*domain.WebhookDelivery, error)
	ListDeliveries(ctx context.Context, webhookID string, limit, offset int) ([]*domain.WebhookDelivery, int64, error)
	GetPendingDeliveries(ctx context.Context, limit int) ([]*domain.WebhookDelivery, error)
	UpdateDelivery(ctx context.Context, delivery *domain.WebhookDelivery) (*domain.WebhookDelivery, error)
	
	// Event operations
	CreateEvent(ctx context.Context, event *domain.WebhookEvent) (*domain.WebhookEvent, error)
	UpdateEvent(ctx context.Context, event *domain.WebhookEvent) (*domain.WebhookEvent, error)
	ListEvents(ctx context.Context, webhookID string, limit, offset int) ([]*domain.WebhookEvent, int64, error)
	
	// Analytics operations
	GetStats(ctx context.Context) (*domain.WebhookStats, error)
}
