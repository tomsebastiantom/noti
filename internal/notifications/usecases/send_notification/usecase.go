package sendnotification

import (
	"context"
	"fmt"

	"getnoti.com/internal/notifications/domain"
	"getnoti.com/internal/notifications/repos"
	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/providers/services"
	"getnoti.com/internal/shared/utils"
	"getnoti.com/internal/templates/services"
	"getnoti.com/internal/tenants/services"
	"getnoti.com/pkg/cache"
)

type SendNotificationUseCase struct {
	tenantService          *tenants.TenantService
	providerService        *providers.ProviderService
	templateService        *templates.TemplateService
	notificationRepository repository.NotificationRepository
	preferencesCache       *cache.GenericCache
}

func NewSendNotificationUseCase(tenantService *tenants.TenantService, providerService *providers.ProviderService, templateService *templates.TemplateService, notificationRepository repository.NotificationRepository, preferencesCache *cache.GenericCache) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		tenantService:          tenantService,
		providerService:        providerService,
		templateService:        templateService,
		notificationRepository: notificationRepository,
		preferencesCache:       preferencesCache,
	}
}

func (u *SendNotificationUseCase) Execute(ctx context.Context, req SendNotificationRequest) SendNotificationResponse {
	providerID, err := u.getProviderID(ctx, req, u.preferencesCache)
	if err != nil {
		return SendNotificationResponse{
			Status: "failed",
			Error:  "failed to get provider ID: " + err.Error(),
		}
	}

	notification, err := u.createNotification(ctx, req, providerID)
	if err != nil {
		return SendNotificationResponse{
			Status: "failed",
			Error:  "notification creation failed: " + err.Error(),
		}
	}

	content, err := u.templateService.GetContent(ctx, notification.TemplateID, notification.Variables)
	if err != nil {
		return SendNotificationResponse{
			ID:     notification.ID,
			Status: "failed",
			Error:  "failed to get template content: " + err.Error(),
		}
	}

	sendReq := dtos.SendNotificationRequest{
		Sender:     req.TenantID,
		Receiver:   req.UserID,
		Channel:    req.Channel,
		Content:    content,
		ProviderID: providerID,
	}

	sendResp := u.providerService.SendNotification(ctx, req.TenantID, providerID, sendReq)
	if !sendResp.Success {
		return SendNotificationResponse{
			ID:     notification.ID,
			Status: "failed",
			Error:  "notification sending failed: " + sendResp.Message, // Detailed error message
		}

	}
	return SendNotificationResponse{
		ID:     notification.ID,
		Status: "queued",
	}
}

func (u *SendNotificationUseCase) getProviderID(ctx context.Context, req SendNotificationRequest, preferencesCache *cache.GenericCache) (string, error) {
	if req.ProviderID != "" {
		return req.ProviderID, nil
	}

	cacheKey := fmt.Sprintf("preferences:%s:%s", req.TenantID, req.Channel)

	// Try to get from cache first
	if cachedPrefs, found := preferencesCache.Get(cacheKey); found {
		preferences := cachedPrefs.(map[string]string)
		return u.extractProviderIDFromPreferences(preferences)
	}

	// If not in cache, fetch from tenant service
	preferences, err := u.tenantService.GetPreferences(ctx, req.TenantID, req.Channel)
	if err != nil {
		return "", fmt.Errorf("failed to fetch preferences: %w", err)
	}

	// Cache the preferences
	preferencesCache.Set(cacheKey, preferences, 1)

	return u.extractProviderIDFromPreferences(preferences)
}

func (u *SendNotificationUseCase) extractProviderIDFromPreferences(preferences map[string]string) (string, error) {
	providerID, exists := preferences["ProviderID"]
	if !exists {
		return "", fmt.Errorf("provider ID not found in preferences")
	}

	enabled, exists := preferences["Enabled"]
	if !exists || enabled != "true" {
		return "", fmt.Errorf("channel not enabled or preference not found")
	}

	return providerID, nil
}

func (u *SendNotificationUseCase) createNotification(ctx context.Context, req SendNotificationRequest, providerID string) (*domain.Notification, error) {
	variables := make([]domain.TemplateVariable, len(req.Variables))
	for i, v := range req.Variables {
		variables[i] = domain.TemplateVariable{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	ID, err := utils.GenerateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}

	notification := &domain.Notification{
		ID:         ID,
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		Type:       req.Type,
		Channel:    req.Channel,
		TemplateID: req.TemplateID,
		Content:    req.Content,
		ProviderID: providerID,
		Variables:  variables,
	}

	err = u.notificationRepository.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}
