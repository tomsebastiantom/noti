package sendnotification

import (
	"context"
	"fmt"
	"sort"

	"getnoti.com/internal/notifications/domain"
	"getnoti.com/internal/notifications/repos"
	providerDomain "getnoti.com/internal/providers/domain"
	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/providers/repos"
	"getnoti.com/internal/providers/services"
	"getnoti.com/internal/shared/utils"
	"getnoti.com/internal/templates/services"

	"getnoti.com/pkg/cache"
)

type SendNotificationUseCase struct {
	providerService        *providers.ProviderService
	templateService        *templates.TemplateService
	providerRepo           repos.ProviderRepository
	notificationRepository repository.NotificationRepository
	preferencesCache       *cache.GenericCache
}

func NewSendNotificationUseCase(providerService *providers.ProviderService, templateService *templates.TemplateService, providerRepo repos.ProviderRepository, notificationRepository repository.NotificationRepository, preferencesCache *cache.GenericCache) *SendNotificationUseCase {
	return &SendNotificationUseCase{
		providerService:        providerService,
		templateService:        templateService,
		providerRepo:           providerRepo,
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
		providers := cachedPrefs.([]*providerDomain.Provider)
		return u.extractProviderIDFromProviders(providers)
	}

	// If not in cache, fetch from provider Repo
	providers, err := u.providerRepo.GetProviders(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch providers: %w", err)
	}

	// Filter providers for the requested channel
	channelProviders := u.filterProvidersByChannel(providers, req.Channel)

	// Sort providers by priority (lowest number = highest priority)
	sort.Slice(channelProviders, func(i, j int) bool {
		return channelProviders[i].Channels[req.Channel].Priority < channelProviders[j].Channels[req.Channel].Priority
	})

	// Cache the providers
	preferencesCache.Set(cacheKey, channelProviders, 1)

	return u.extractProviderIDFromProviders(channelProviders)
}

func (u *SendNotificationUseCase) extractProviderIDFromProviders(providers []*providerDomain.Provider) (string, error) {
	for _, provider := range providers {
		if provider.Enabled {
			return provider.ID, nil
		}
	}
	return "", fmt.Errorf("no enabled provider found")
}

func (u *SendNotificationUseCase) filterProvidersByChannel(providers []*providerDomain.Provider, channel string) []*providerDomain.Provider {
	var channelProviders []*providerDomain.Provider
	for _, provider := range providers {
		if _, ok := provider.Channels[channel]; ok {
			channelProviders = append(channelProviders, provider)
		}
	}
	return channelProviders
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
