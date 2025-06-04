package domain

import (
	"time"
)

// Webhook represents a webhook configuration
type Webhook struct {
	ID          string            `json:"id"`
	TenantID    string            `json:"tenant_id"`
	URL         string            `json:"url"`
	Secret      string            `json:"-"` // Never expose in JSON
	Events      []string          `json:"events"`
	Headers     map[string]string `json:"headers"`
	IsActive    bool              `json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	LastUsedAt  *time.Time        `json:"last_used_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string     `json:"id"`
	WebhookID    string     `json:"webhook_id"`
	TenantID     string     `json:"tenant_id"`
	EventType    string     `json:"event_type"`
	EventID      string     `json:"event_id"`
	Payload      string     `json:"payload"`
	StatusCode   int        `json:"status_code"`
	Response     string     `json:"response"`
	AttemptCount int        `json:"attempt_count"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	CreatedAt    time.Time  `json:"created_at"`
	NextRetryAt  *time.Time `json:"next_retry_at"`
}

// IncrementAttempt increments the attempt count for the delivery
func (wd *WebhookDelivery) IncrementAttempt() {
	wd.AttemptCount++
}

// MarkFailed marks the delivery as failed with error details
func (wd *WebhookDelivery) MarkFailed(statusCode int, errorMessage string, nextRetryAt *time.Time) {
	wd.StatusCode = statusCode
	wd.Response = errorMessage
	wd.NextRetryAt = nextRetryAt
}

// MarkDelivered marks the delivery as successfully delivered
func (wd *WebhookDelivery) MarkDelivered(statusCode int, response string) {
	wd.StatusCode = statusCode
	wd.Response = response
	now := time.Now()
	wd.DeliveredAt = &now
}

// WebhookEvent represents a webhook event record
type WebhookEvent struct {
	ID         string      `json:"id"`
	WebhookID  string      `json:"webhook_id"`
	TenantID   string      `json:"tenant_id"`
	EventType  string      `json:"event_type"`
	EventData  interface{} `json:"event_data"`
	DeliveryID *string     `json:"delivery_id"`
	CreatedAt  time.Time   `json:"created_at"`
}

// WebhookStats represents webhook statistics for a tenant
type WebhookStats struct {
	TotalWebhooks        int64 `json:"total_webhooks"`
	ActiveWebhooks       int64 `json:"active_webhooks"`
	TotalDeliveries      int64 `json:"total_deliveries"`
	SuccessfulDeliveries int64 `json:"successful_deliveries"`
	FailedDeliveries     int64 `json:"failed_deliveries"`
	PendingDeliveries    int64 `json:"pending_deliveries"`
	AverageResponseTime  int64 `json:"average_response_time_ms"`
}
