package sendnotification

import (
	"context"
	"fmt"

	"getnoti.com/internal/notifications/domain"
	notificationRepos "getnoti.com/internal/notifications/repos"
	providerDomain "getnoti.com/internal/providers/domain"
	"getnoti.com/internal/providers/dtos"
	providerRepos "getnoti.com/internal/providers/repos"
	providerServices "getnoti.com/internal/providers/services"
	"getnoti.com/internal/shared/utils"
	templateServices "getnoti.com/internal/templates/services"

	"getnoti.com/pkg/cache"
)

type SendNotificationUseCase struct {
	providerService        *providerServices.ProviderService
	templateService        *templateServices.TemplateService
	providerRepo           providerRepos.ProviderRepository
	notificationRepository notificationRepos.NotificationRepository
	preferencesCache       *cache.GenericCache
}

func NewSendNotificationUseCase(
	providerService *providerServices.ProviderService, 
	templateService *templateServices.TemplateService, 
	providerRepo providerRepos.ProviderRepository, 
	notificationRepository notificationRepos.NotificationRepository, 
	preferencesCache *cache.GenericCache,
) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		providerService:        providerService,
		templateService:        templateService,
		providerRepo:           providerRepo,
		notificationRepository: notificationRepository,
		preferencesCache:       preferencesCache,
	}
}

func (u *SendNotificationUseCase) Execute(ctx context.Context, req SendNotificationRequest) (SendNotificationResponse, error) {
	providerID, err := u.getProviderID(ctx, req, u.preferencesCache)
	if err != nil {
		return SendNotificationResponse{
			Status: "failed",
			Error:  "failed to get provider ID: " + err.Error(),
		}, err
	}

	notification, err := u.createNotification(ctx, req, providerID)
	if err != nil {
		return SendNotificationResponse{
			Status: "failed",
			Error:  "notification creation failed: " + err.Error(),
		}, err
	}
	content, err := u.templateService.GetContent(ctx, req.TenantID, notification.TemplateID, notification.Variables)
	if err != nil {
		return SendNotificationResponse{
			ID:     notification.ID,
			Status: "failed",
			Error:  "failed to get template content: " + err.Error(),
		}, err
	}

	sendReq := dtos.SendNotificationRequest{
		Sender:     req.TenantID,
		Receiver:   req.UserID,
		TenantID:   req.TenantID,
		Channel:    req.Channel,
		Content:    content,
		ProviderID: providerID,
	}
	sendResp := u.providerService.DispatchNotification(ctx, req.TenantID, providerID, sendReq)
	if !sendResp.Success {
		return SendNotificationResponse{
			ID:     notification.ID,
			Status: "failed",
			Error:  "notification sending failed: " + sendResp.Message,
		}, fmt.Errorf("notification sending failed: %s", sendResp.Message)
	}
	
	return SendNotificationResponse{
		ID:     notification.ID,
		Status: "queued",
	}, nil
}

func (u *SendNotificationUseCase) getProviderID(ctx context.Context, req SendNotificationRequest, preferencesCache *cache.GenericCache) (string, error) {
	if req.ProviderID != "" {
		return req.ProviderID, nil
	}

	cacheKey := fmt.Sprintf("preferences:%s:%s", req.TenantID, req.Channel)

	// Try to get from cache first
	if cachedProvider, found := preferencesCache.Get(cacheKey); found {
		provider := cachedProvider.(*providerDomain.Provider)
		return provider.ID, nil
	}

	// If not in cache, fetch from provider Repo
	provider, err := u.providerRepo.GetProviderByChannel(ctx, req.Channel)
	if err != nil {
		return "", fmt.Errorf("failed to fetch provider: %w", err)
	}

	if provider == nil {
		return "", fmt.Errorf("no provider found for channel %s", req.Channel)
	}

	// Cache the provider
	preferencesCache.Set(cacheKey, provider, 1)

	return provider.ID, nil
}

func (u *SendNotificationUseCase) createNotification(ctx context.Context, req SendNotificationRequest, providerID string) (*domain.Notification, error) {
	variables := make([]domain.TemplateVariable, len(req.Variables))
	for i, v := range req.Variables {
		variables[i] = domain.TemplateVariable{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	ID := utils.GenerateUUID()

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

	err := u.notificationRepository.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// TODO: Implement fallback mechanism to try the next available provider if the current one fails to send the notification.
