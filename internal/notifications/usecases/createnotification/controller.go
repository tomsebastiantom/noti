
package createnotification

import (
    "context"
    "getnoti.com/internal/notifications/dtos"
)

type CreateNotificationController struct {
    useCase CreateNotificationUseCase
}

func NewCreateNotificationController(useCase CreateNotificationUseCase) *CreateNotificationController {
    return &CreateNotificationController{useCase: useCase}
}

func (c *CreateNotificationController) CreateNotification(ctx context.Context, req dto.NotificationDTO) dto.NotificationDTO {
    return c.useCase.Execute(ctx, req)
}
