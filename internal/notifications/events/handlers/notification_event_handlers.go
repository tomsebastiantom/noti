package handlers

import (
	"context"
	"fmt"

	notificationEvents "getnoti.com/internal/notifications/events"
	"getnoti.com/internal/shared/events"
	"getnoti.com/pkg/logger"
)

// NotificationEventHandlers handles notification domain events
type NotificationEventHandlers struct {
	logger              logger.Logger
}

// NewNotificationEventHandlers creates a new notification event handlers instance
func NewNotificationEventHandlers(
	logger logger.Logger,
) *NotificationEventHandlers {
	return &NotificationEventHandlers{
		logger:              logger,
	}
}

// HandleNotificationCreated processes notification created events
func (h *NotificationEventHandlers) HandleNotificationCreated(ctx context.Context, event events.DomainEvent) error {
	notificationEvent, ok := event.(*notificationEvents.NotificationCreatedEvent)
	if !ok {
		h.logger.Error("Invalid event type for NotificationCreated handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "expected_type", Value: "NotificationCreatedEvent"},
			logger.Field{Key: "actual_type", Value: fmt.Sprintf("%T", event)})
		return fmt.Errorf("invalid event type: expected NotificationCreatedEvent, got %T", event)
	}

	h.logger.Info("Processing notification created event",
		logger.Field{Key: "event_id", Value: notificationEvent.GetEventID()},
		logger.Field{Key: "notification_id", Value: notificationEvent.NotificationID},
		logger.Field{Key: "tenant_id", Value: notificationEvent.GetTenantID()})

	// Domain-specific business logic for notification creation
	// This could trigger:
	// - Template validation
	// - User preference checks
	// - Provider selection logic
	// - Delivery scheduling

	return nil
}

// HandleNotificationDeliveryRequested processes delivery request events
func (h *NotificationEventHandlers) HandleNotificationDeliveryRequested(ctx context.Context, event events.DomainEvent) error {
	deliveryEvent, ok := event.(*notificationEvents.NotificationDeliveryRequestedEvent)
	if !ok {
		h.logger.Error("Invalid event type for NotificationDeliveryRequested handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected NotificationDeliveryRequestedEvent, got %T", event)
	}
	h.logger.Info("Processing notification delivery requested event",
		logger.Field{Key: "event_id", Value: deliveryEvent.GetEventID()},
		logger.Field{Key: "notification_id", Value: deliveryEvent.NotificationID},
		logger.Field{Key: "provider_id", Value: deliveryEvent.ProviderID},
		logger.Field{Key: "tenant_id", Value: deliveryEvent.GetTenantID()})

	// Domain-specific delivery logic
	// This could trigger:
	// - Provider authentication
	// - Rate limiting checks
	// - Retry scheduling
	// - Circuit breaker checks

	return nil
}

// HandleNotificationDelivered processes successful delivery events
func (h *NotificationEventHandlers) HandleNotificationDelivered(ctx context.Context, event events.DomainEvent) error {
	deliveredEvent, ok := event.(*notificationEvents.NotificationDeliveredEvent)
	if !ok {
		h.logger.Error("Invalid event type for NotificationDelivered handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected NotificationDeliveredEvent, got %T", event)
	}

	h.logger.Info("Processing notification delivered event",
		logger.Field{Key: "event_id", Value: deliveredEvent.GetEventID()},
		logger.Field{Key: "notification_id", Value: deliveredEvent.NotificationID},
		logger.Field{Key: "provider_id", Value: deliveredEvent.ProviderID},
		logger.Field{Key: "tenant_id", Value: deliveredEvent.GetTenantID()})

	// Domain-specific success logic
	// This could trigger:
	// - Analytics updates
	// - Webhook notifications
	// - Status updates
	// - Cleanup tasks

	return nil
}

// HandleNotificationFailed processes failed delivery events
func (h *NotificationEventHandlers) HandleNotificationFailed(ctx context.Context, event events.DomainEvent) error {
	failedEvent, ok := event.(*notificationEvents.NotificationFailedEvent)
	if !ok {
		h.logger.Error("Invalid event type for NotificationFailed handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected NotificationFailedEvent, got %T", event)
	}

	h.logger.Error("Processing notification failed event",
		logger.Field{Key: "event_id", Value: failedEvent.GetEventID()},
		logger.Field{Key: "notification_id", Value: failedEvent.NotificationID},
		logger.Field{Key: "provider_id", Value: failedEvent.ProviderID},
		logger.Field{Key: "error", Value: failedEvent.ErrorMessage},
		logger.Field{Key: "retry_count", Value: failedEvent.RetryCount},
		logger.Field{Key: "tenant_id", Value: failedEvent.GetTenantID()})

	// Domain-specific failure logic
	// This could trigger:
	// - Retry scheduling
	// - Fallback provider selection
	// - Alerting
	// - Circuit breaker updates

	return nil
}

// GetHandlerMethods returns a map of event types to handler methods for registration
func (h *NotificationEventHandlers) GetHandlerMethods() map[string]func(context.Context, events.DomainEvent) error {
	return map[string]func(context.Context, events.DomainEvent) error{
		notificationEvents.NotificationCreatedEventType:          h.HandleNotificationCreated,
		notificationEvents.NotificationDeliveryRequestedEventType: h.HandleNotificationDeliveryRequested,
		notificationEvents.NotificationSentEventType:            h.HandleNotificationDelivered,
		notificationEvents.NotificationFailedEventType:          h.HandleNotificationFailed,
	}
}
