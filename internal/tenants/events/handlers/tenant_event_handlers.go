package handlers

import (
	"context"
	"fmt"

	"getnoti.com/internal/shared/events"
	tenantEvents "getnoti.com/internal/tenants/events"
	"getnoti.com/pkg/logger"
)

// TenantEventHandlers handles tenant and user preference domain events
type TenantEventHandlers struct {
	logger                  logger.Logger
}

// NewTenantEventHandlers creates a new tenant event handlers instance
func NewTenantEventHandlers(
	logger logger.Logger,
) *TenantEventHandlers {
	return &TenantEventHandlers{
		logger:               logger,
	}
}

// HandleUserPreferenceUpdated processes user preference updated events
func (h *TenantEventHandlers) HandleUserPreferenceUpdated(ctx context.Context, event events.DomainEvent) error {
	preferenceEvent, ok := event.(*tenantEvents.UserPreferenceUpdatedEvent)
	if !ok {
		h.logger.Error("Invalid event type for UserPreferenceUpdated handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()},
			logger.Field{Key: "expected_type", Value: "UserPreferenceUpdatedEvent"},
			logger.Field{Key: "actual_type", Value: fmt.Sprintf("%T", event)})
		return fmt.Errorf("invalid event type: expected UserPreferenceUpdatedEvent, got %T", event)
	}

	h.logger.Info("Processing user preference updated event",
		logger.Field{Key: "event_id", Value: preferenceEvent.GetEventID()},
		logger.Field{Key: "user_id", Value: preferenceEvent.UserID},
		logger.Field{Key: "preference_type", Value: preferenceEvent.PreferenceType},
		logger.Field{Key: "tenant_id", Value: preferenceEvent.GetTenantID()})

	// Domain-specific business logic for preference updates
	// This could trigger:
	// - Cache invalidation
	// - Notification routing updates
	// - Provider preference changes
	// - Analytics updates

	return nil
}

// HandleTenantConfigurationChanged processes tenant configuration change events
func (h *TenantEventHandlers) HandleTenantConfigurationChanged(ctx context.Context, event events.DomainEvent) error {
	configEvent, ok := event.(*tenantEvents.TenantConfigurationUpdatedEvent)
	if !ok {
		h.logger.Error("Invalid event type for TenantConfigurationChanged handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected TenantConfigurationUpdatedEvent, got %T", event)
	}

	h.logger.Info("Processing tenant configuration changed event",
		logger.Field{Key: "event_id", Value: configEvent.GetEventID()},
		logger.Field{Key: "config_key", Value: configEvent.ConfigKey},
		logger.Field{Key: "tenant_id", Value: configEvent.GetTenantID()})

	// Domain-specific configuration change logic
	// This could trigger:
	// - Cache invalidation
	// - Service reconfiguration
	// - Provider updates
	// - Security policy updates

	return nil
}

// HandleUserCreated processes user creation events
func (h *TenantEventHandlers) HandleUserCreated(ctx context.Context, event events.DomainEvent) error {
	userEvent, ok := event.(*tenantEvents.UserCreatedEvent)
	if !ok {
		h.logger.Error("Invalid event type for UserCreated handler",
			logger.Field{Key: "event_id", Value: event.GetEventID()})
		return fmt.Errorf("invalid event type: expected UserCreatedEvent, got %T", event)
	}

	h.logger.Info("Processing user created event",
		logger.Field{Key: "event_id", Value: userEvent.GetEventID()},
		logger.Field{Key: "user_id", Value: userEvent.UserID},
		logger.Field{Key: "email", Value: userEvent.Email},
		logger.Field{Key: "tenant_id", Value: userEvent.GetTenantID()})

	// Domain-specific user creation logic
	// This could trigger:
	// - Default preference setup
	// - Welcome notification
	// - Analytics tracking
	// - Audit logging

	return nil
}

// GetHandlerMethods returns a map of event types to handler methods for registration
func (h *TenantEventHandlers) GetHandlerMethods() map[string]func(context.Context, events.DomainEvent) error {
	return map[string]func(context.Context, events.DomainEvent) error{
		tenantEvents.UserCreatedEventType:                  h.HandleUserCreated,
		tenantEvents.UserPreferenceUpdatedEventType:       h.HandleUserPreferenceUpdated,
		tenantEvents.TenantConfigurationUpdatedEventType:  h.HandleTenantConfigurationChanged,
	}
}
