package providers

import (
    "context"
    "getnoti.com/internal/providers/dtos"
)

type Provider interface {
    SendNotification(ctx context.Context, req dtos.SendNotificationRequest) dtos.SendNotificationResponse
}