package sendnotification

import (
	"context"

	"getnoti.com/internal/notifications/domain"
	dto "getnoti.com/internal/notifications/dtos"
	repository "getnoti.com/internal/notifications/repos"
)

type SendNotificationUseCase interface {
	Execute(ctx context.Context, req dto.NotificationDTO) dto.NotificationDTO
}

type sendNotificationUseCase struct {
	notificationRepository repository.NotificationRepository
}

func NewSendNotificationUseCase(notificationRepository repository.NotificationRepository) SendNotificationUseCase {
	return &sendNotificationUseCase{notificationRepository: notificationRepository}
}

func (u *sendNotificationUseCase) Execute(ctx context.Context, req dto.NotificationDTO) dto.NotificationDTO {
	variables := make([]domain.TemplateVariable, len(req.Variables))
	for i, v := range req.Variables {
		variables[i] = domain.TemplateVariable{
			Key:   v.Key,
			Value: v.Value,
		}
	}

	notification := &domain.Notification{
		ID:         req.ID,
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		Type:       req.Type,
		Channel:    req.Channel,
		TemplateID: req.TemplateID,
		Status:     req.Status,
		Content:    req.Content,
		Variables:  variables,
	}

	err := u.notificationRepository.CreateNotification(ctx, notification)
	if err != nil {
		return dto.NotificationDTO{Error: "notification creation failed"}
	}
	//we will send the notification if immediate

	return dto.NotificationDTO{
		ID:         notification.ID,
		TenantID:   notification.TenantID,
		UserID:     notification.UserID,
		Type:       notification.Type,
		Channel:    notification.Channel,
		TemplateID: notification.TemplateID,
		Status:     notification.Status,
		Content:    notification.Content,
		Variables:  req.Variables,
	}
}
