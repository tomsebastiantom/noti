package providers

import (
    "context"
    "getnoti.com/internal/providers/dtos"
    "getnoti.com/internal/providers/infra/providers"
    "getnoti.com/pkg/queue"
    "getnoti.com/pkg/workerpool"
)

type ProviderService struct {
    factory             *providers.ProviderFactory
    notificationManager *NotificationManager
}

func NewProviderService(factory *providers.ProviderFactory, queue queue.Queue, wpm *workerpool.WorkerPoolManager) *ProviderService {
    return &ProviderService{
        factory:             factory,
        notificationManager: NewNotificationManager(queue, factory, wpm),
    }
}

func (s *ProviderService) DispatchNotification(ctx context.Context, tenantID string, providerID string, req dtos.SendNotificationRequest) dtos.SendNotificationResponse {
    req.ProviderID = providerID
    req.Sender = tenantID

    err := s.notificationManager.DispatchNotification(ctx, req)
    if err != nil {
        return dtos.SendNotificationResponse{Success: false, Message: "Failed to send notification"}
    }

    return dtos.SendNotificationResponse{Success: true, Message: "Notification sent successfully"}
}

func (s *ProviderService) Shutdown() {
    s.notificationManager.Shutdown()
}

func (s *ProviderService) GetPreferences(ctx context.Context,l string,lo string) map[string]string {
    return map[string]string{
        "df": "dsd",
        "gh": "dd",
    }
}


