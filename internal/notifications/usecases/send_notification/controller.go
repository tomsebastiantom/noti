package sendnotification

import (
	"context"


)

type SendNotificationController struct {
	useCase *SendNotificationUseCase
}

func NewSendNotificationController(useCase *SendNotificationUseCase) *SendNotificationController {
	return &SendNotificationController{useCase: useCase}
}

func (c *SendNotificationController) SendNotification(ctx context.Context, req SendNotificationRequest) SendNotificationResponse {
	return c.useCase.Execute(ctx, req)
}
