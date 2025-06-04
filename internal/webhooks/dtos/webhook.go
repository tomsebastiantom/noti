package dto

import (
	"time"
)

// CreateWebhookRequest represents the request to create a new webhook
type CreateWebhookRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=255"`
	URL            string            `json:"url" validate:"required,url"`
	Events         []string          `json:"events" validate:"required,min=1"`
	Headers        map[string]string `json:"headers,omitempty"`
	RetryCount     *int              `json:"retry_count,omitempty" validate:"omitempty,min=0,max=10"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty" validate:"omitempty,min=1,max=300"`
	IsActive       *bool             `json:"is_active,omitempty"`
}

// UpdateWebhookRequest represents the request to update an existing webhook
type UpdateWebhookRequest struct {
	Name           *string           `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	URL            *string           `json:"url,omitempty" validate:"omitempty,url"`
	Events         []string          `json:"events,omitempty" validate:"omitempty,min=1"`
	Headers        map[string]string `json:"headers,omitempty"`
	RetryCount     *int              `json:"retry_count,omitempty" validate:"omitempty,min=0,max=10"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty" validate:"omitempty,min=1,max=300"`
	IsActive       *bool             `json:"is_active,omitempty"`
}

// WebhookResponse represents a webhook in API responses
type WebhookResponse struct {
	ID             string            `json:"id"`
	TenantID       string            `json:"tenant_id"`
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	Events         []string          `json:"events"`
	Headers        map[string]string `json:"headers"`
	IsActive       bool              `json:"is_active"`
	RetryCount     int               `json:"retry_count"`
	TimeoutSeconds int               `json:"timeout_seconds"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// WebhookListResponse represents a list of webhooks with pagination
type WebhookListResponse struct {
	Webhooks   []WebhookResponse `json:"webhooks"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// WebhookDeliveryResponse represents a webhook delivery in API responses
type WebhookDeliveryResponse struct {
	ID                 string    `json:"id"`
	WebhookID          string    `json:"webhook_id"`
	TenantID           string    `json:"tenant_id"`
	EventType          string    `json:"event_type"`
	EventData          any       `json:"event_data"`
	Status             string    `json:"status"`
	AttemptCount       int       `json:"attempt_count"`
	MaxAttempts        int       `json:"max_attempts"`
	NextRetryAt        *time.Time `json:"next_retry_at,omitempty"`
	ResponseStatusCode *int      `json:"response_status_code,omitempty"`
	ResponseBody       *string   `json:"response_body,omitempty"`
	ErrorMessage       *string   `json:"error_message,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// WebhookDeliveryListResponse represents a list of webhook deliveries with pagination
type WebhookDeliveryListResponse struct {
	Deliveries []WebhookDeliveryResponse `json:"deliveries"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

// WebhookEventResponse represents a webhook event in API responses
type WebhookEventResponse struct {
	ID         string    `json:"id"`
	WebhookID  string    `json:"webhook_id"`
	TenantID   string    `json:"tenant_id"`
	EventType  string    `json:"event_type"`
	EventData  any       `json:"event_data"`
	DeliveryID *string   `json:"delivery_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// WebhookEventListResponse represents a list of webhook events with pagination
type WebhookEventListResponse struct {
	Events     []WebhookEventResponse `json:"events"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// WebhookStatsResponse represents webhook statistics
type WebhookStatsResponse struct {
	TotalWebhooks     int64                  `json:"total_webhooks"`
	ActiveWebhooks    int64                  `json:"active_webhooks"`
	TotalDeliveries   int64                  `json:"total_deliveries"`
	SuccessfulDeliveries int64              `json:"successful_deliveries"`
	FailedDeliveries  int64                  `json:"failed_deliveries"`
	PendingDeliveries int64                  `json:"pending_deliveries"`
	DeliverysByStatus map[string]int64       `json:"deliveries_by_status"`
	RecentEvents      []WebhookEventResponse `json:"recent_events"`
}

// RegenerateSecretResponse represents the response after regenerating webhook secret
type RegenerateSecretResponse struct {
	NewSecret string `json:"new_secret"`
	Message   string `json:"message"`
}

// TestWebhookRequest represents a request to test webhook delivery
type TestWebhookRequest struct {
	EventType string `json:"event_type" validate:"required"`
	EventData any    `json:"event_data" validate:"required"`
}

// TestWebhookResponse represents the response of a webhook test
type TestWebhookResponse struct {
	Success            bool    `json:"success"`
	StatusCode         int     `json:"status_code"`
	ResponseBody       string  `json:"response_body"`
	ResponseTime       string  `json:"response_time"`
	ErrorMessage       *string `json:"error_message,omitempty"`
}
