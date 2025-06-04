package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"getnoti.com/internal/webhooks/domain"
	"getnoti.com/internal/webhooks/repos"
	"getnoti.com/pkg/db"
	"getnoti.com/pkg/logger"
	"getnoti.com/pkg/queue"
	"getnoti.com/pkg/webhook"
)

// WebhookService handles webhook-related business operations
type WebhookService struct {
	connectionManager *db.Manager
	queueManager      *queue.QueueManager
	securityManager   *webhook.SecurityManager
	sender            *webhook.Sender
	logger            logger.Logger
	repositoryFactory interface {
		GetWebhookRepositoryForTenant(tenantID string) (repos.WebhookRepository, error)
	}
}

// NewWebhookService creates a new webhook service
func NewWebhookService(
	connectionManager *db.Manager,
	queueManager *queue.QueueManager,
	securityManager *webhook.SecurityManager,
	sender *webhook.Sender,
	logger logger.Logger,
	repositoryFactory interface {
		GetWebhookRepositoryForTenant(tenantID string) (repos.WebhookRepository, error)
	},
) *WebhookService {
	return &WebhookService{
		connectionManager: connectionManager,
		queueManager:      queueManager,
		securityManager:   securityManager,
		sender:            sender,
		logger:            logger,
		repositoryFactory: repositoryFactory,
	}
}

// CreateWebhook creates a new webhook configuration
func (s *WebhookService) CreateWebhook(ctx context.Context, tenantID string, webhook *domain.Webhook) (*domain.Webhook, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook repository: %w", err)
	}

	// Generate secret if not provided
	if webhook.Secret == "" {
		secret, err := s.securityManager.GenerateSecret()
		if err != nil {
			s.logger.Error("Failed to generate webhook secret", 
				logger.String("tenant_id", tenantID), 
				logger.Err(err))
			return nil, fmt.Errorf("failed to generate webhook secret: %w", err)
		}
		webhook.Secret = secret
	}

	// Set tenant ID
	webhook.TenantID = tenantID

	// Basic validation - check required fields
	if webhook.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	if len(webhook.Events) == 0 {
		return nil, fmt.Errorf("webhook events are required")
	}

	// Create webhook in database
	createdWebhook, err := webhookRepo.CreateWebhook(ctx, webhook)
	if err != nil {
		s.logger.Error("Failed to create webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_url", webhook.URL), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	s.logger.Info("Webhook created successfully", 
		logger.String("tenant_id", tenantID), 
		logger.String("webhook_id", createdWebhook.ID), 
		logger.String("webhook_url", createdWebhook.URL))
	
	return createdWebhook, nil
}

// GetWebhook retrieves a webhook by ID
func (s *WebhookService) GetWebhook(ctx context.Context, tenantID, webhookID string) (*domain.Webhook, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant",
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook repository: %w", err)
	}

	webhook, err := webhookRepo.GetWebhookByID(ctx, webhookID)
	if err != nil {
		s.logger.Error("Failed to get webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	return webhook, nil
}

// UpdateWebhook updates an existing webhook
func (s *WebhookService) UpdateWebhook(ctx context.Context, tenantID, webhookID string, updates *domain.Webhook) (*domain.Webhook, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook repository: %w", err)
	}

	// Get existing webhook
	existingWebhook, err := webhookRepo.GetWebhookByID(ctx, webhookID)
	if err != nil {
		s.logger.Error("Failed to get existing webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get existing webhook: %w", err)
	}

	// Apply updates
	if updates.URL != "" {
		existingWebhook.URL = updates.URL
	}
	if len(updates.Events) > 0 {
		existingWebhook.Events = updates.Events
	}
	if updates.Headers != nil {
		existingWebhook.Headers = updates.Headers
	}
	existingWebhook.IsActive = updates.IsActive
	existingWebhook.UpdatedAt = time.Now()

	// Basic validation
	if existingWebhook.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	if len(existingWebhook.Events) == 0 {
		return nil, fmt.Errorf("webhook events are required")
	}

	updatedWebhook, err := webhookRepo.UpdateWebhook(ctx, existingWebhook)
	if err != nil {
		s.logger.Error("Failed to update webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	s.logger.Info("Webhook updated successfully", 
		logger.String("tenant_id", tenantID), 
		logger.String("webhook_id", webhookID))
	
	return updatedWebhook, nil
}

// DeleteWebhook deletes a webhook
func (s *WebhookService) DeleteWebhook(ctx context.Context, tenantID, webhookID string) error {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return fmt.Errorf("failed to get webhook repository: %w", err)
	}

	err = webhookRepo.DeleteWebhook(ctx, webhookID)
	if err != nil {
		s.logger.Error("Failed to delete webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	s.logger.Info("Webhook deleted successfully", 
		logger.String("tenant_id", tenantID), 
		logger.String("webhook_id", webhookID))
	
	return nil
}

// ListWebhooks retrieves webhooks with pagination
func (s *WebhookService) ListWebhooks(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Webhook, int64, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, 0, fmt.Errorf("failed to get webhook repository: %w", err)
	}

	webhooks, total, err := webhookRepo.ListWebhooks(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list webhooks", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, 0, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return webhooks, total, nil
}

// RegenerateSecret generates a new secret for a webhook
func (s *WebhookService) RegenerateSecret(ctx context.Context, tenantID, webhookID string) (string, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return "", fmt.Errorf("failed to get webhook repository: %w", err)
	}

	// Generate new secret
	newSecret, err := s.securityManager.GenerateSecret()
	if err != nil {
		s.logger.Error("Failed to generate new secret", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return "", fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Update webhook with new secret
	webhook, err := webhookRepo.GetWebhookByID(ctx, webhookID)
	if err != nil {
		s.logger.Error("Failed to get webhook for secret regeneration", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return "", fmt.Errorf("failed to get webhook: %w", err)
	}

	webhook.Secret = newSecret
	webhook.UpdatedAt = time.Now()
	
	_, err = webhookRepo.UpdateWebhook(ctx, webhook)
	if err != nil {
		s.logger.Error("Failed to update webhook with new secret", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return "", fmt.Errorf("failed to update webhook secret: %w", err)
	}

	s.logger.Info("Webhook secret regenerated successfully", 
		logger.String("tenant_id", tenantID), 
		logger.String("webhook_id", webhookID))
	
	return newSecret, nil
}

// DispatchEvent dispatches an event to all relevant webhooks
func (s *WebhookService) DispatchEvent(ctx context.Context, tenantID, eventType string, eventData interface{}) error {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return fmt.Errorf("failed to get webhook repository: %w", err)
	}

	// Get active webhooks that listen to this event type
	webhooks, err := webhookRepo.GetWebhooksByEventType(ctx, eventType)
	if err != nil {
		s.logger.Error("Failed to get webhooks for event type", 
			logger.String("tenant_id", tenantID), 
			logger.String("event_type", eventType), 
			logger.Err(err))
		return fmt.Errorf("failed to get webhooks for event type: %w", err)
	}

	if len(webhooks) == 0 {
		s.logger.Debug("No webhooks found for event type", 
			logger.String("tenant_id", tenantID), 
			logger.String("event_type", eventType))
		return nil
	}

	// Dispatch to each webhook
	for _, webhook := range webhooks {
		if !webhook.IsActive {
			continue
		}

		// Create webhook event record
		event := &domain.WebhookEvent{
			WebhookID: webhook.ID,
			TenantID:  tenantID,
			EventType: eventType,
			EventData: eventData,
			CreatedAt: time.Now(),
		}

		// Create event record
		createdEvent, err := webhookRepo.CreateEvent(ctx, event)
		if err != nil {
			s.logger.Error("Failed to create webhook event", 
				logger.String("tenant_id", tenantID), 
				logger.String("webhook_id", webhook.ID), 
				logger.String("event_type", eventType), 
				logger.Err(err))
			continue
		}

		// Serialize event data for delivery
		payloadBytes, err := json.Marshal(eventData)
		if err != nil {
			s.logger.Error("Failed to serialize event data", 
				logger.String("tenant_id", tenantID), 
				logger.String("webhook_id", webhook.ID), 
				logger.Err(err))
			continue
		}

		// Create delivery record
		delivery := &domain.WebhookDelivery{
			WebhookID:    webhook.ID,
			TenantID:     tenantID,
			EventType:    eventType,
			EventID:      createdEvent.ID,
			Payload:      string(payloadBytes),
			AttemptCount: 0,
			CreatedAt:    time.Now(),
		}

		createdDelivery, err := webhookRepo.CreateDelivery(ctx, delivery)
		if err != nil {
			s.logger.Error("Failed to create webhook delivery", 
				logger.String("tenant_id", tenantID), 
				logger.String("webhook_id", webhook.ID), 
				logger.Err(err))
			continue
		}

		// Link event to delivery
		createdEvent.DeliveryID = &createdDelivery.ID
		_, err = webhookRepo.UpdateEvent(ctx, createdEvent)
		if err != nil {
			s.logger.Warn("Failed to link event to delivery", 
				logger.String("tenant_id", tenantID), 
				logger.String("event_id", createdEvent.ID), 
				logger.String("delivery_id", createdDelivery.ID), 
				logger.Err(err))
		}

		// Queue the delivery for processing
		if err := s.queueWebhookDelivery(ctx, tenantID, createdDelivery.ID); err != nil {
			s.logger.Error("Failed to queue webhook delivery", 
				logger.String("tenant_id", tenantID), 
				logger.String("delivery_id", createdDelivery.ID), 
				logger.Err(err))
		}
	}

	s.logger.Info("Event dispatched to webhooks", 
		logger.String("tenant_id", tenantID), 
		logger.String("event_type", eventType), 
		logger.Int("webhook_count", len(webhooks)))
	
	return nil
}

// TestWebhook sends a test event to a webhook
func (s *WebhookService) TestWebhook(ctx context.Context, tenantID, webhookID, eventType string, eventData interface{}) (*domain.WebhookDelivery, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook repository: %w", err)
	}

	// Get webhook
	webhook, err := webhookRepo.GetWebhookByID(ctx, webhookID)
	if err != nil {
		s.logger.Error("Failed to get webhook for testing", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	// Serialize test data
	payloadBytes, err := json.Marshal(eventData)
	if err != nil {
		s.logger.Error("Failed to serialize test event data", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to serialize test event data: %w", err)
	}

	// Create test delivery
	delivery := &domain.WebhookDelivery{
		WebhookID:    webhook.ID,
		TenantID:     tenantID,
		EventType:    eventType,
		EventID:      "test-event-id",
		Payload:      string(payloadBytes),
		AttemptCount: 0,
		CreatedAt:    time.Now(),
	}

	createdDelivery, err := webhookRepo.CreateDelivery(ctx, delivery)
	if err != nil {
		s.logger.Error("Failed to create test delivery", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to create test delivery: %w", err)
	}

	// Send webhook immediately (synchronous for test)
	err = s.sendWebhook(ctx, webhook, createdDelivery)
	if err != nil {
		s.logger.Error("Failed to send test webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhookID), 
			logger.String("delivery_id", createdDelivery.ID), 
			logger.Err(err))
	}

	// Get updated delivery with response
	updatedDelivery, err := webhookRepo.GetDeliveryByID(ctx, createdDelivery.ID)
	if err != nil {
		s.logger.Error("Failed to get updated test delivery", 
			logger.String("tenant_id", tenantID), 
			logger.String("delivery_id", createdDelivery.ID), 
			logger.Err(err))
		return createdDelivery, nil
	}

	return updatedDelivery, nil
}

// GetWebhookStats retrieves webhook statistics for a tenant
func (s *WebhookService) GetWebhookStats(ctx context.Context, tenantID string) (*domain.WebhookStats, error) {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook repository: %w", err)
	}

	stats, err := webhookRepo.GetStats(ctx)
	if err != nil {
		s.logger.Error("Failed to get webhook stats", 
			logger.String("tenant_id", tenantID), 
			logger.Err(err))
		return nil, fmt.Errorf("failed to get webhook stats: %w", err)
	}

	return stats, nil
}

// queueWebhookDelivery queues a webhook delivery for background processing
func (s *WebhookService) queueWebhookDelivery(ctx context.Context, tenantID, deliveryID string) error {
	// Create queue message for webhook delivery
	message := map[string]interface{}{
		"tenant_id":   tenantID,
		"delivery_id": deliveryID,
		"type":        "webhook_delivery",
	}

	// Queue the message (implementation depends on QueueManager interface)
	s.logger.Info("Queueing webhook delivery", 
		logger.String("tenant_id", tenantID), 
		logger.String("delivery_id", deliveryID))

	// TODO: Implement actual queueing when QueueManager interface is defined
	_ = message // Prevent unused variable error
	return nil
}

// sendWebhook sends a webhook delivery using the webhook sender
func (s *WebhookService) sendWebhook(ctx context.Context, wh *domain.Webhook, delivery *domain.WebhookDelivery) error {
	// Create webhook event for sending
	webhookEvent := webhook.WebhookEvent{
		TenantID:  wh.TenantID,
		EventType: delivery.EventType,
		EventID:   delivery.EventID,
		Timestamp: delivery.CreatedAt,
		Data:      map[string]interface{}{"payload": delivery.Payload},
	}

	// Create webhook config from domain webhook
	webhookConfig := webhook.WebhookConfig{
		ID:        wh.ID,
		TenantID:  wh.TenantID,
		URL:       wh.URL,
		Secret:    wh.Secret,
		Events:    wh.Events,
		Headers:   wh.Headers,
		IsActive:  wh.IsActive,
		CreatedAt: wh.CreatedAt,
		UpdatedAt: wh.UpdatedAt,
	}

	// Send webhook using the sender
	result, err := s.sender.SendEvent(ctx, webhookConfig, webhookEvent)
	if err != nil {
		s.logger.Error("Failed to send webhook", 
			logger.String("webhook_id", wh.ID), 
			logger.String("url", wh.URL),
			logger.Err(err))
		
		// Update delivery record with failure
		webhookRepo, connErr := s.repositoryFactory.GetWebhookRepositoryForTenant(wh.TenantID)
		if connErr != nil {
			s.logger.Error("Failed to get webhook repository for delivery update", 
				logger.String("tenant_id", wh.TenantID), 
				logger.Err(connErr))
			return fmt.Errorf("failed to get webhook repository: %w", connErr)
		}
		delivery.IncrementAttempt()
		delivery.MarkFailed(0, err.Error(), nil)

		_, updateErr := webhookRepo.UpdateDelivery(ctx, delivery)
		if updateErr != nil {
			s.logger.Error("Failed to update delivery", 
				logger.String("tenant_id", wh.TenantID), 
				logger.String("delivery_id", delivery.ID), 
				logger.Err(updateErr))
		}

		return err
	}

	// Update delivery record with success
	webhookRepo, connErr := s.repositoryFactory.GetWebhookRepositoryForTenant(wh.TenantID)
	if connErr != nil {
		s.logger.Error("Failed to get webhook repository for delivery update", 
			logger.String("tenant_id", wh.TenantID), 
			logger.Err(connErr))
		return fmt.Errorf("failed to get webhook repository: %w", connErr)
	}

	delivery.IncrementAttempt()
	delivery.MarkDelivered(result.StatusCode, result.Response)

	_, updateErr := webhookRepo.UpdateDelivery(ctx, delivery)
	if updateErr != nil {
		s.logger.Error("Failed to update delivery", 
			logger.String("tenant_id", wh.TenantID), 
			logger.String("delivery_id", delivery.ID), 
			logger.Err(updateErr))
	}

	s.logger.Info("Webhook sent successfully", 
		logger.String("webhook_id", wh.ID), 
		logger.String("url", wh.URL),
		logger.Int("status_code", result.StatusCode))

	return nil
}

// ProcessWebhookDelivery processes a queued webhook delivery
func (s *WebhookService) ProcessWebhookDelivery(ctx context.Context, tenantID, deliveryID string) error {
	// Get tenant-specific repository
	webhookRepo, err := s.repositoryFactory.GetWebhookRepositoryForTenant(tenantID)
	if err != nil {
		s.logger.Error("Failed to get webhook repository for tenant", 
			logger.String("tenant_id", tenantID), 
			logger.String("delivery_id", deliveryID),
			logger.Err(err))
		return fmt.Errorf("failed to get webhook repository: %w", err)
	}

	// Get delivery record
	delivery, err := webhookRepo.GetDeliveryByID(ctx, deliveryID)
	if err != nil {
		s.logger.Error("Failed to get delivery for processing", 
			logger.String("tenant_id", tenantID), 
			logger.String("delivery_id", deliveryID),
			logger.Err(err))
		return fmt.Errorf("failed to get delivery: %w", err)
	}

	// Get webhook configuration
	webhook, err := webhookRepo.GetWebhookByID(ctx, delivery.WebhookID)
	if err != nil {
		s.logger.Error("Failed to get webhook for delivery processing", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", delivery.WebhookID),
			logger.String("delivery_id", deliveryID),
			logger.Err(err))
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	// Check if webhook is still active
	if !webhook.IsActive {
		s.logger.Warn("Skipping delivery for inactive webhook", 
			logger.String("tenant_id", tenantID), 
			logger.String("webhook_id", webhook.ID),
			logger.String("delivery_id", deliveryID))
		return nil
	}

	// Send the webhook
	return s.sendWebhook(ctx, webhook, delivery)
}
