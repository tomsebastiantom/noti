package create_webhook

import (
	dto "getnoti.com/internal/webhooks/dtos"
)

// CreateWebhookRequest represents the request to create a webhook
type CreateWebhookRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=255"`
	URL            string            `json:"url" validate:"required,url"`
	Events         []string          `json:"events" validate:"required,min=1"`
	Headers        map[string]string `json:"headers,omitempty"`
	RetryCount     *int              `json:"retry_count,omitempty" validate:"omitempty,min=0,max=10"`
	TimeoutSeconds *int              `json:"timeout_seconds,omitempty" validate:"omitempty,min=1,max=300"`
	IsActive       *bool             `json:"is_active,omitempty"`
}

// CreateWebhookResponse represents the response after creating a webhook
type CreateWebhookResponse struct {
	Webhook dto.WebhookResponse `json:"webhook"`
	Secret  string              `json:"secret"` // Only included in creation response
}
