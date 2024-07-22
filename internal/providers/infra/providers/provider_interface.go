package providers

import (
    "context"
    "getnoti.com/internal/providers/dtos"
)

type Provider interface {
    CreateClient(ctx context.Context, credentials map[string]interface{}) error
    SendNotification(ctx context.Context, req dtos.SendNotificationRequest) dtos.SendNotificationResponse
}
