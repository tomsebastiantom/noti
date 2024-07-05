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
	providerRepository     queue.repo
}

func NewSendNotificationUseCase(notificationRepository repository.NotificationRepository, providerRepository repository.ProviderRepository) SendNotificationUseCase {
	return &sendNotificationUseCase{
		notificationRepository: notificationRepository,
		providerRepository:     providerRepository,
	}
}

func (u *sendNotificationUseCase) Execute(ctx context.Context, req dto.NotificationDTO) dto.NotificationDTO {
	// Fetch default provider ID if not provided
	providerID := req.ProviderID
	if providerID == "" {
		defaultProviderID, err := u.providerRepository.GetDefaultProviderID(ctx, req.TenantID)
		if err != nil {
			return dto.NotificationDTO{
				Status: "failed",
				Error:  "failed to fetch default provider ID: " + err.Error(),
			}
		}
		providerID = defaultProviderID
	}

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
		ProviderID: providerID, // Set the provider ID
		Variables:  variables,
	}

	err := u.notificationRepository.CreateNotification(ctx, notification)
	if err != nil {
		return dto.NotificationDTO{
			ID:     notification.ID,
			Status: "failed",
			Error:  "notification creation failed: " + err.Error(), // Detailed error message
		}
	}

	// Assuming there's a function to send the notification immediately
	sendErr := u.sendNotification(notification)
	if sendErr != nil {
		return dto.NotificationDTO{
			ID:     notification.ID,
			Status: "failed",
			Error:  "notification sending failed: " + sendErr.Error(), // Detailed error message
		}
	}

	return dto.NotificationDTO{
		ID:     notification.ID,
		Status: "sent",
	}
}

// Mock function to represent sending the notification
func (u *sendNotificationUseCase) sendNotification(notification *domain.Notification) error {
	// Implement the actual sending logic here
	// Return an error if sending fails
	return nil
}
