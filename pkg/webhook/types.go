package webhook

import (
	"time"
)

// WebhookConfig represents a webhook configuration stored in tenant DB
type WebhookConfig struct {
	ID          string            `json:"id" db:"id"`
	TenantID    string            `json:"tenant_id" db:"tenant_id"`
	URL         string            `json:"url" db:"url"`
	Secret      string            `json:"-" db:"secret"` // Never expose in JSON
	Events      []string          `json:"events" db:"events"`
	Headers     map[string]string `json:"headers" db:"headers"`
	IsActive    bool              `json:"is_active" db:"is_active"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	LastUsedAt  *time.Time        `json:"last_used_at" db:"last_used_at"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID           string     `json:"id" db:"id"`
	WebhookID    string     `json:"webhook_id" db:"webhook_id"`
	TenantID     string     `json:"tenant_id" db:"tenant_id"`
	EventType    string     `json:"event_type" db:"event_type"`
	EventID      string     `json:"event_id" db:"event_id"`
	Payload      string     `json:"payload" db:"payload"`
	StatusCode   int        `json:"status_code" db:"status_code"`
	Response     string     `json:"response" db:"response"`
	AttemptCount int        `json:"attempt_count" db:"attempt_count"`
	DeliveredAt  *time.Time `json:"delivered_at" db:"delivered_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	NextRetryAt  *time.Time `json:"next_retry_at" db:"next_retry_at"`
}

// WebhookEvent represents an event to be sent via webhook
type WebhookEvent struct {
	TenantID  string                 `json:"tenant_id"`
	EventType string                 `json:"event_type"`
	EventID   string                 `json:"event_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// OutgoingWebhook represents a webhook to be sent
type OutgoingWebhook struct {
	WebhookID string            `json:"webhook_id"`
	URL       string            `json:"url"`
	Payload   []byte            `json:"payload"`
	Headers   map[string]string `json:"headers"`
	TenantID  string            `json:"tenant_id"`
	EventType string            `json:"event_type"`
	EventID   string            `json:"event_id"`
	Secret    string            `json:"-"` // Never expose in JSON
}

// IncomingWebhook represents a received webhook
type IncomingWebhook struct {
	TenantID    string                 `json:"tenant_id"`
	EventType   string                 `json:"event_type"`
	EventID     string                 `json:"event_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
	Signature   string                 `json:"-"` // From headers
}

// WebhookCreateRequest represents a request to create a webhook
type WebhookCreateRequest struct {
	URL     string            `json:"url" validate:"required,url"`
	Events  []string          `json:"events" validate:"required,min=1"`
	Headers map[string]string `json:"headers,omitempty"`
}

// WebhookUpdateRequest represents a request to update a webhook
type WebhookUpdateRequest struct {
	URL      *string            `json:"url,omitempty" validate:"omitempty,url"`
	Events   []string          `json:"events,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	IsActive *bool             `json:"is_active,omitempty"`
}
