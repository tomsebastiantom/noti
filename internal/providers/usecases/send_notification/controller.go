package sendnotification

import (
    "context"

    dto "getnoti.com/internal/providers/dtos"
)

type SendNotificationController struct {
    useCase SendNotificationUseCase
}

func NewSendNotificationController(useCase SendNotificationUseCase) *SendNotificationController {
    return &SendNotificationController{useCase: useCase}
}

func (c *SendNotificationController) SendNotification(ctx context.Context, req dto.SendNotificationRequest) dto.SendNotificationResponse {
    return c.useCase.Execute(ctx, req)
}
