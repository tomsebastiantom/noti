package providers

import (
    "context"
    "getnoti.com/internal/providers/dtos"
    "getnoti.com/internal/providers/infra/providers"
    "getnoti.com/pkg/logger"
    "getnoti.com/pkg/queue"
)

type ProviderService struct {
    factory             *providers.ProviderFactory
    notificationManager *NotificationManager
    logger              logger.Logger
}

func NewProviderService(factory *providers.ProviderFactory, queue queue.Queue, log logger.Logger) *ProviderService {
    return &ProviderService{
        factory:             factory,
        notificationManager: NewNotificationManager(queue, factory, log),
        logger:              log,
    }
}

func (s *ProviderService) SendNotification(ctx context.Context, tenantID string, providerID string, req dtos.SendNotificationRequest) dtos.SendNotificationResponse {
    // Update the request with the providerID
    req.ProviderID = providerID
    req.Sender = tenantID

    err := s.notificationManager.SendNotification(ctx, req)
    if err != nil {
        s.logger.Error("Failed to send notification: %v", err)
        return dtos.SendNotificationResponse{Success: false, Message: "Failed to send notification"}
    }

    // Return a successful response
    return dtos.SendNotificationResponse{Success: true, Message: "Notification sent successfully"}
}

func (s *ProviderService) Shutdown() {
    s.notificationManager.Shutdown()
}
