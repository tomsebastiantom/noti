package create_webhook

import (
	"context"

	"getnoti.com/internal/webhooks/domain"
	dto "getnoti.com/internal/webhooks/dtos"
	"getnoti.com/internal/webhooks/services"
	"getnoti.com/pkg/logger"
)

// UseCase handles webhook creation business logic
type UseCase struct {
	webhookService *services.WebhookService
	logger         logger.Logger
}

// NewUseCase creates a new create webhook use case
func NewUseCase(
	webhookService *services.WebhookService,
	logger logger.Logger,
) *UseCase {
	return &UseCase{
		webhookService: webhookService,
		logger:         logger,
	}
}

// Execute creates a new webhook
func (uc *UseCase) Execute(ctx context.Context, tenantID string, req *CreateWebhookRequest) (*CreateWebhookResponse, error) {	// Simple validation - check required fields
	if req.Name == "" {
		uc.logger.Warn("Missing webhook name", logger.String("tenant_id", tenantID))
		return nil, NewValidationError("name is required")
	}
	if req.URL == "" {
		uc.logger.Warn("Missing webhook URL", logger.String("tenant_id", tenantID))
		return nil, NewValidationError("url is required")
	}
	if len(req.Events) == 0 {
		uc.logger.Warn("Missing webhook events", logger.String("tenant_id", tenantID))
		return nil, NewValidationError("events are required")
	}

	// Convert DTO to domain model
	webhook := &domain.Webhook{
		TenantID: tenantID,
		URL:      req.URL,
		Events:   req.Events,
		Headers:  req.Headers,
		IsActive: true, // Default to active
	}

	// Set default headers if nil
	if webhook.Headers == nil {
		webhook.Headers = make(map[string]string)
	}

	// Create webhook
	createdWebhook, err := uc.webhookService.CreateWebhook(ctx, tenantID, webhook)
	if err != nil {
		uc.logger.Error("Failed to create webhook", 
			logger.String("tenant_id", tenantID),
			logger.String("webhook_url", req.URL),
			logger.String("error", err.Error()))
		return nil, err
	}
	// Build response
	response := &CreateWebhookResponse{
		Webhook: dto.WebhookResponse{
			ID:             createdWebhook.ID,
			TenantID:       createdWebhook.TenantID,
			Name:           req.Name,  // Use the original request name
			URL:            createdWebhook.URL,
			Events:         createdWebhook.Events,
			Headers:        createdWebhook.Headers,
			IsActive:       createdWebhook.IsActive,
			RetryCount:     3,  // Default values since domain doesn't have these
			TimeoutSeconds: 30,
			CreatedAt:      createdWebhook.CreatedAt,
			UpdatedAt:      createdWebhook.UpdatedAt,
		},
		Secret: createdWebhook.Secret, // Include secret only in creation response
	}

	uc.logger.Info("Webhook created successfully", 
		logger.String("tenant_id", tenantID),
		logger.String("webhook_id", createdWebhook.ID),
		logger.String("webhook_url", createdWebhook.URL))

	return response, nil
}
