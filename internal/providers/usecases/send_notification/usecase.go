package sendnotification

import (
	"context"

	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/providers/infra/providers"
)

type SendNotificationUseCase interface {
	Execute(ctx context.Context, req dtos.SendNotificationRequest) dtos.SendNotificationResponse
}

type sendNotificationUseCase struct {
	provider providers.Provider
}

func NewSendNotificationUseCase(provider providers.Provider) SendNotificationUseCase {
	return &sendNotificationUseCase{
		provider: provider,
	}
}

func (uc *sendNotificationUseCase) Execute(ctx context.Context, req dtos.SendNotificationRequest) dtos.SendNotificationResponse {
	return uc.provider.SendNotification(ctx, req)
}
