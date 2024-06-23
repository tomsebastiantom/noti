package sendnotification

import (
	"context"

	dto "getnoti.com/internal/notifications/dtos"
)

type SendNotificationController struct {
	useCase SendNotificationUseCase
}

func NewSendNotificationController(useCase SendNotificationUseCase) *SendNotificationController {
	return &SendNotificationController{useCase: useCase}
}

func (c *SendNotificationController) SendNotification(ctx context.Context, req dto.NotificationDTO) dto.NotificationDTO {
	return c.useCase.Execute(ctx, req)
}
