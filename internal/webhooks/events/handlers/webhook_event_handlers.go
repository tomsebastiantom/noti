package handlers

import (
	"context"
	"fmt"

	"getnoti.com/internal/shared/events"
	webhookEvents "getnoti.com/internal/webhooks/events"
	"getnoti.com/pkg/logger"
)

// WebhookEventHandlers handles webhook domain events
type WebhookEventHandlers struct {
	logger         logger.Logger
}

// NewWebhookEventHandlers creates a new webhook event handlers instance
func NewWebhookEventHandlers(
	logger logger.Logger,
) *WebhookEventHandlers {
	return &WebhookEventHandlers{
		logger:        logger,
	}
}

// HandleWebhookDispatchRequested processes webhook dispatch request events
func (h *WebhookEventHandlers) HandleWebhookDispatchRequested(ctx context.Context, event events.DomainEvent) error {
	webhookEvent, ok := event.(*webhookEvents.WebhookDispatchRequestedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WebhookDispatchRequested handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "expected_type", Value: "WebhookDispatchRequestedEvent"},
			logger.Field{Key: "actual_type", Value: fmt.Sprintf("%T", event)})
		return fmt.Errorf("invalid event type: expected WebhookDispatchRequestedEvent, got %T", event)
	}

	h.logger.Info("Processing webhook dispatch requested event",
		logger.Field{Key: "event_id", Value: webhookEvent.GetEventID()},
		logger.Field{Key: "webhook_id", Value: webhookEvent.WebhookID},
		logger.Field{Key: "target_url", Value: webhookEvent.TargetURL},
		logger.Field{Key: "event_type", Value: webhookEvent.EventType},
		logger.Field{Key: "tenant_id", Value: webhookEvent.GetTenantID()})

	// Domain-specific webhook dispatch logic
	// This could trigger:
	// - Webhook validation
	// - Security checks
	// - Rate limiting
	// - Retry scheduling

	return nil
}

// HandleWebhookDelivered processes successful webhook delivery events
func (h *WebhookEventHandlers) HandleWebhookDelivered(ctx context.Context, event events.DomainEvent) error {
	deliveredEvent, ok := event.(*webhookEvents.WebhookDeliveredEvent)
	if !ok {
		h.logger.Error("Invalid event type for WebhookDelivered handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected WebhookDeliveredEvent, got %T", event)
	}
	h.logger.Info("Processing webhook delivered event",
		logger.Field{Key: "event_id", Value: deliveredEvent.GetEventID()},
		logger.Field{Key: "webhook_id", Value: deliveredEvent.WebhookID},
		logger.Field{Key: "target_url", Value: deliveredEvent.TargetURL},
		logger.Field{Key: "response_status", Value: deliveredEvent.ResponseStatus},
		logger.Field{Key: "tenant_id", Value: deliveredEvent.GetTenantID()})

	// Domain-specific webhook success logic
	// This could trigger:
	// - Analytics updates
	// - Status tracking
	// - Cleanup tasks
	// - Success notifications

	return nil
}

// HandleWebhookFailed processes failed webhook delivery events
func (h *WebhookEventHandlers) HandleWebhookFailed(ctx context.Context, event events.DomainEvent) error {
	failedEvent, ok := event.(*webhookEvents.WebhookFailedEvent)
	if !ok {
		h.logger.Error("Invalid event type for WebhookFailed handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected WebhookFailedEvent, got %T", event)
	}

	h.logger.Error("Processing webhook failed event",
		logger.Field{Key: "event_id", Value: failedEvent.GetEventID()},
		logger.Field{Key: "webhook_id", Value: failedEvent.WebhookID},
		logger.Field{Key: "target_url", Value: failedEvent.TargetURL},
		logger.Field{Key: "error", Value: failedEvent.ErrorMessage},
		logger.Field{Key: "retry_count", Value: failedEvent.RetryCount},
		logger.Field{Key: "tenant_id", Value: failedEvent.GetTenantID()})

	// Domain-specific webhook failure logic
	// This could trigger:
	// - Retry scheduling
	// - Circuit breaker updates
	// - Alerting
	// - Fallback mechanisms

	return nil
}

// GetHandlerMethods returns a map of event types to handler methods for registration
func (h *WebhookEventHandlers) GetHandlerMethods() map[string]func(context.Context, events.DomainEvent) error {
	return map[string]func(context.Context, events.DomainEvent) error{
		webhookEvents.WebhookDispatchRequestedEventType: h.HandleWebhookDispatchRequested,
		webhookEvents.WebhookDeliveredEventType:         h.HandleWebhookDelivered,
		webhookEvents.WebhookFailedEventType:            h.HandleWebhookFailed,
	}
}
