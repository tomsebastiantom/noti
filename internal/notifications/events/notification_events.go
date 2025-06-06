package events

import (
	"getnoti.com/internal/shared/events"
)

// Notification Event Types
const (
	NotificationCreatedEventType          = "notification.created"
	NotificationDeliveryRequestedEventType = "notification.delivery_requested"
	NotificationSentEventType            = "notification.sent"
	NotificationFailedEventType          = "notification.failed"
	NotificationUpdatedEventType         = "notification.updated"
	NotificationDeletedEventType         = "notification.deleted"
)

// Notification Domain Events

// NotificationCreatedEvent is published when a new notification is created
type NotificationCreatedEvent struct {
	*events.BaseDomainEvent
	NotificationID   string                 `json:"notification_id"`
	UserID           string                 `json:"user_id"`
	TemplateID       string                 `json:"template_id,omitempty"`
	Channel          string                 `json:"channel"`
	Priority         string                 `json:"priority"`
	ScheduledFor     *string                `json:"scheduled_for,omitempty"`
	NotificationData map[string]interface{} `json:"notification_data"`
}

// NewNotificationCreatedEvent creates a new notification created event
func NewNotificationCreatedEvent(
	notificationID, userID, tenantID, templateID, channel, priority string,
	scheduledFor *string,
	notificationData map[string]interface{},
) *NotificationCreatedEvent {
	payload := map[string]interface{}{
		"notification_id":   notificationID,
		"user_id":          userID,
		"template_id":      templateID,
		"channel":          channel,
		"priority":         priority,
		"scheduled_for":    scheduledFor,
		"notification_data": notificationData,
	}
	
	return &NotificationCreatedEvent{
		BaseDomainEvent:  events.NewBaseDomainEvent("notification.created", notificationID, tenantID, payload),
		NotificationID:   notificationID,
		UserID:           userID,
		TemplateID:       templateID,
		Channel:          channel,
		Priority:         priority,
		ScheduledFor:     scheduledFor,
		NotificationData: notificationData,
	}
}

// NotificationDeliveryRequestedEvent is published when notification delivery is requested
type NotificationDeliveryRequestedEvent struct {
	*events.BaseDomainEvent
	NotificationID string                 `json:"notification_id"`
	UserID         string                 `json:"user_id"`
	ProviderID     string                 `json:"provider_id"`
	Channel        string                 `json:"channel"`
	Priority       string                 `json:"priority"`
	RetryCount     int                    `json:"retry_count"`
	DeliveryData   map[string]interface{} `json:"delivery_data"`
}

// NewNotificationDeliveryRequestedEvent creates a new delivery requested event
func NewNotificationDeliveryRequestedEvent(
	notificationID, userID, tenantID, providerID, channel, priority string,
	retryCount int,
	deliveryData map[string]interface{},
) *NotificationDeliveryRequestedEvent {
	payload := map[string]interface{}{
		"notification_id": notificationID,
		"user_id":        userID,
		"provider_id":    providerID,
		"channel":        channel,
		"priority":       priority,
		"retry_count":    retryCount,
		"delivery_data":  deliveryData,
	}
	
	return &NotificationDeliveryRequestedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("notification.delivery.requested", notificationID, tenantID, payload),
		NotificationID:  notificationID,
		UserID:          userID,
		ProviderID:      providerID,
		Channel:         channel,
		Priority:        priority,
		RetryCount:      retryCount,
		DeliveryData:    deliveryData,
	}
}

// NotificationDeliveredEvent is published when notification is successfully delivered
type NotificationDeliveredEvent struct {
	*events.BaseDomainEvent
	NotificationID string                 `json:"notification_id"`
	UserID         string                 `json:"user_id"`
	ProviderID     string                 `json:"provider_id"`
	Channel        string                 `json:"channel"`
	DeliveredAt    string                 `json:"delivered_at"`
	DeliveryInfo   map[string]interface{} `json:"delivery_info"`
}

// NewNotificationDeliveredEvent creates a new notification delivered event
func NewNotificationDeliveredEvent(
	notificationID, userID, tenantID, providerID, channel, deliveredAt string,
	deliveryInfo map[string]interface{},
) *NotificationDeliveredEvent {
	payload := map[string]interface{}{
		"notification_id": notificationID,
		"user_id":        userID,
		"provider_id":    providerID,
		"channel":        channel,
		"delivered_at":   deliveredAt,
		"delivery_info":  deliveryInfo,
	}
	
	return &NotificationDeliveredEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("notification.delivered", notificationID, tenantID, payload),
		NotificationID:  notificationID,
		UserID:          userID,
		ProviderID:      providerID,
		Channel:         channel,
		DeliveredAt:     deliveredAt,
		DeliveryInfo:    deliveryInfo,
	}
}

// NotificationFailedEvent is published when notification delivery fails
type NotificationFailedEvent struct {
	*events.BaseDomainEvent
	NotificationID string                 `json:"notification_id"`
	UserID         string                 `json:"user_id"`
	ProviderID     string                 `json:"provider_id"`
	Channel        string                 `json:"channel"`
	FailedAt       string                 `json:"failed_at"`
	ErrorMessage   string                 `json:"error_message"`
	RetryCount     int                    `json:"retry_count"`
	WillRetry      bool                   `json:"will_retry"`
	ErrorDetails   map[string]interface{} `json:"error_details"`
}

// NewNotificationFailedEvent creates a new notification failed event
func NewNotificationFailedEvent(
	notificationID, userID, tenantID, providerID, channel, failedAt, errorMessage string,
	retryCount int, willRetry bool,
	errorDetails map[string]interface{},
) *NotificationFailedEvent {
	payload := map[string]interface{}{
		"notification_id": notificationID,
		"user_id":        userID,
		"provider_id":    providerID,
		"channel":        channel,
		"failed_at":      failedAt,
		"error_message":  errorMessage,
		"retry_count":    retryCount,
		"will_retry":     willRetry,
		"error_details":  errorDetails,
	}
	
	return &NotificationFailedEvent{
		BaseDomainEvent: events.NewBaseDomainEvent("notification.failed", notificationID, tenantID, payload),
		NotificationID:  notificationID,
		UserID:          userID,
		ProviderID:      providerID,
		Channel:         channel,
		FailedAt:        failedAt,
		ErrorMessage:    errorMessage,
		RetryCount:      retryCount,
		WillRetry:       willRetry,
		ErrorDetails:    errorDetails,
	}
}
