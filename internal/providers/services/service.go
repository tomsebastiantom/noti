package providers

import (
    "context"
    "getnoti.com/internal/providers/infra/providers"
    "getnoti.com/internal/providers/dtos"
)

type ProviderService struct {
    factory *providers.ProviderFactory
}

func NewProviderService(factory *providers.ProviderFactory) *ProviderService {
    return &ProviderService{
        factory: factory,
    }
}

func (s *ProviderService) SendNotification(ctx context.Context, tenantID string, providerID string, req dtos.SendNotificationRequest) dtos.SendNotificationResponse {
    provider := s.factory.GetProvider(providerID, tenantID, req.Channel)
    if provider == nil {
        return dtos.SendNotificationResponse{Success: false, Message: "Unsupported provider"}
    }

    return provider.SendNotification(ctx, req)
}
