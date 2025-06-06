package events

import (
	"getnoti.com/internal/shared/events"
)

// Webhook Event Types
const (
	WebhookDispatchRequestedEventType = "webhook.dispatch_requested"
	WebhookDeliveredEventType        = "webhook.delivered"
	WebhookFailedEventType           = "webhook.failed"
	WebhookCreatedEventType          = "webhook.created"
	WebhookUpdatedEventType          = "webhook.updated"
	WebhookDeletedEventType          = "webhook.deleted"
)

// Webhook Domain Events

// WebhookDispatchRequestedEvent is published when a webhook needs to be dispatched
type WebhookDispatchRequestedEvent struct {
	*events.BaseDomainEvent
	WebhookID    string                 `json:"webhook_id"`
	EventType    string                 `json:"event_type"`
	TargetURL    string                 `json:"target_url"`
	HTTPMethod   string                 `json:"http_method"`
	Headers      map[string]string      `json:"headers"`
	PayloadData  map[string]interface{} `json:"payload_data"`
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
	ScheduledFor *string                `json:"scheduled_for,omitempty"`
}

// NewWebhookDispatchRequestedEvent creates a new webhook dispatch requested event
func NewWebhookDispatchRequestedEvent(
	webhookID, tenantID, eventType, targetURL, httpMethod string,
	headers map[string]string,
	payloadData map[string]interface{},
	retryCount, maxRetries int,
	scheduledFor *string,
) *WebhookDispatchRequestedEvent {
	payload := map[string]interface{}{
		"webhook_id":    webhookID,
		"event_type":    eventType,
		"target_url":    targetURL,
		"http_method":   httpMethod,
		"headers":       headers,
		"payload_data":  payloadData,
		"retry_count":   retryCount,
		"max_retries":   maxRetries,
		"scheduled_for": scheduledFor,
	}
	
	return &WebhookDispatchRequestedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("webhook.dispatch.requested", webhookID, tenantID, payload),
		WebhookID:       webhookID,
		EventType:       eventType,
		TargetURL:       targetURL,
		HTTPMethod:      httpMethod,
		Headers:         headers,
		PayloadData:     payloadData,
		RetryCount:      retryCount,
		MaxRetries:      maxRetries,
		ScheduledFor:    scheduledFor,
	}
}

// WebhookDeliveredEvent is published when webhook is successfully delivered
type WebhookDeliveredEvent struct {
	*events.BaseDomainEvent
	WebhookID      string                 `json:"webhook_id"`
	EventType      string                 `json:"event_type"`
	TargetURL      string                 `json:"target_url"`
	DeliveredAt    string                 `json:"delivered_at"`
	ResponseStatus int                    `json:"response_status"`
	ResponseTime   int64                  `json:"response_time_ms"`
	RetryCount     int                    `json:"retry_count"`
	DeliveryInfo   map[string]interface{} `json:"delivery_info"`
}

// NewWebhookDeliveredEvent creates a new webhook delivered event
func NewWebhookDeliveredEvent(
	webhookID, tenantID, eventType, targetURL, deliveredAt string,
	responseStatus int, responseTime int64, retryCount int,
	deliveryInfo map[string]interface{},
) *WebhookDeliveredEvent {
	payload := map[string]interface{}{
		"webhook_id":      webhookID,
		"event_type":      eventType,
		"target_url":      targetURL,
		"delivered_at":    deliveredAt,
		"response_status": responseStatus,
		"response_time":   responseTime,
		"retry_count":     retryCount,
		"delivery_info":   deliveryInfo,
	}
	
	return &WebhookDeliveredEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("webhook.delivered", webhookID, tenantID, payload),
		WebhookID:       webhookID,
		EventType:       eventType,
		TargetURL:       targetURL,
		DeliveredAt:     deliveredAt,
		ResponseStatus:  responseStatus,
		ResponseTime:    responseTime,
		RetryCount:      retryCount,
		DeliveryInfo:    deliveryInfo,
	}
}

// WebhookFailedEvent is published when webhook delivery fails
type WebhookFailedEvent struct {
	*events.BaseDomainEvent
	WebhookID      string                 `json:"webhook_id"`
	EventType      string                 `json:"event_type"`
	TargetURL      string                 `json:"target_url"`
	FailedAt       string                 `json:"failed_at"`
	ErrorMessage   string                 `json:"error_message"`
	ResponseStatus int                    `json:"response_status"`
	RetryCount     int                    `json:"retry_count"`
	MaxRetries     int                    `json:"max_retries"`
	WillRetry      bool                   `json:"will_retry"`
	NextRetryAt    *string                `json:"next_retry_at,omitempty"`
	ErrorDetails   map[string]interface{} `json:"error_details"`
}

// NewWebhookFailedEvent creates a new webhook failed event
func NewWebhookFailedEvent(
	webhookID, tenantID, eventType, targetURL, failedAt, errorMessage string,
	responseStatus, retryCount, maxRetries int,
	willRetry bool, nextRetryAt *string,
	errorDetails map[string]interface{},
) *WebhookFailedEvent {
	payload := map[string]interface{}{
		"webhook_id":      webhookID,
		"event_type":      eventType,
		"target_url":      targetURL,
		"failed_at":       failedAt,
		"error_message":   errorMessage,
		"response_status": responseStatus,
		"retry_count":     retryCount,
		"max_retries":     maxRetries,
		"will_retry":      willRetry,
		"next_retry_at":   nextRetryAt,
		"error_details":   errorDetails,
	}
	
	return &WebhookFailedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("webhook.failed", webhookID, tenantID, payload),
		WebhookID:       webhookID,
		EventType:       eventType,
		TargetURL:       targetURL,
		FailedAt:        failedAt,
		ErrorMessage:    errorMessage,
		ResponseStatus:  responseStatus,
		RetryCount:      retryCount,
		MaxRetries:      maxRetries,
		WillRetry:       willRetry,
		NextRetryAt:     nextRetryAt,
		ErrorDetails:    errorDetails,
	}
}

// WebhookRetryScheduledEvent is published when webhook retry is scheduled
type WebhookRetryScheduledEvent struct {
	*events.BaseDomainEvent
	WebhookID   string `json:"webhook_id"`
	EventType   string `json:"event_type"`
	TargetURL   string `json:"target_url"`
	RetryCount  int    `json:"retry_count"`
	MaxRetries  int    `json:"max_retries"`
	ScheduledAt string `json:"scheduled_at"`
	RetryDelay  int64  `json:"retry_delay_seconds"`
}

// NewWebhookRetryScheduledEvent creates a new webhook retry scheduled event
func NewWebhookRetryScheduledEvent(
	webhookID, tenantID, eventType, targetURL, scheduledAt string,
	retryCount, maxRetries int, retryDelay int64,
) *WebhookRetryScheduledEvent {
	payload := map[string]interface{}{
		"webhook_id":   webhookID,
		"event_type":   eventType,
		"target_url":   targetURL,
		"retry_count":  retryCount,
		"max_retries":  maxRetries,
		"scheduled_at": scheduledAt,
		"retry_delay":  retryDelay,
	}
	
	return &WebhookRetryScheduledEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("webhook.retry.scheduled", webhookID, tenantID, payload),
		WebhookID:       webhookID,
		EventType:       eventType,
		TargetURL:       targetURL,
		RetryCount:      retryCount,
		MaxRetries:      maxRetries,
		ScheduledAt:     scheduledAt,
		RetryDelay:      retryDelay,
	}
}
