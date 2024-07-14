package sendnotification

import (
	"context"

	"getnoti.com/internal/notifications/domain"
	"getnoti.com/internal/notifications/repos"
	"getnoti.com/internal/providers/services"
	"getnoti.com/internal/providers/dtos"
	"getnoti.com/internal/tenants/services"
	"getnoti.com/internal/shared/utils"
)



type SendNotificationUseCase struct {
    tenantService       *tenants.TenantService
    providerService     *providers.ProviderService
    notificationRepository repository.NotificationRepository
}

func NewSendNotificationUseCase(tenantService *tenants.TenantService, providerService *providers.ProviderService, notificationRepository repository.NotificationRepository) *SendNotificationUseCase {
    return &SendNotificationUseCase{
        tenantService:       tenantService,
        providerService:     providerService,
        notificationRepository: notificationRepository,
    }
}

func (u *SendNotificationUseCase) Execute(ctx context.Context, req SendNotificationRequest) SendNotificationResponse {
    // Fetch default provider ID if not provided
    providerID := req.ProviderID
    if providerID == "" {
        preferences, err := u.tenantService.GetPreferences(ctx, req.TenantID, req.Channel)
        if err != nil {
            return SendNotificationResponse{
                Status: "failed",
                Error:  "failed to fetch preferences: " + err.Error(),
            }
        }
        channelPreference, exists := preferences["ProviderID"]
        if !exists {
            return SendNotificationResponse{
                Status: "failed",
                Error:  "provider ID not found in preferences",
            }
        }
        enabled, exists := preferences["Enabled"]
        if !exists || enabled != "true" {
            return SendNotificationResponse{
                Status: "failed",
                Error:  "channel not enabled or preference not found",
            }
        }
        providerID = channelPreference
    }

    variables := make([]domain.TemplateVariable, len(req.Variables))
    for i, v := range req.Variables {
        variables[i] = domain.TemplateVariable{
            Key:   v.Key,
            Value: v.Value,
        }
    }
    ID, err := utils.GenerateUUID()
    if err != nil {
        return SendNotificationResponse{
            Status: "failed",
            Error:  "notification creation failed: " + err.Error(),
        }
    }
    notification := &domain.Notification{
        ID:         ID,
        TenantID:   req.TenantID,
        UserID:     req.UserID,
        Type:       req.Type,
        Channel:    req.Channel,
        TemplateID: req.TemplateID,
        Content:    req.Content,
        ProviderID: providerID, // Set the provider ID
        Variables:  variables,
    }

    err = u.notificationRepository.CreateNotification(ctx, notification)
    if err != nil {
        return SendNotificationResponse{
            ID:     notification.ID,
            Status: "failed",
            Error:  "notification creation failed: " + err.Error(), // Detailed error message
        }
    }

    // Send the notification immediately
    sendReq := dtos.SendNotificationRequest{
        Sender:    req.TenantID,
        Receiver:  req.UserID,
        Channel:   req.Channel,
        Content:   req.Content,
        ProviderID: providerID,
    }
    sendResp := u.providerService.SendNotification(ctx, req.TenantID, providerID, sendReq)
    if !sendResp.Success {
        return SendNotificationResponse{
            ID:     notification.ID,
            Status: "failed",
            Error:  "notification sending failed: " + sendResp.Message, // Detailed error message
        }
    }
//update notification repo saying it was send sucessfully
    return SendNotificationResponse{
        ID:     notification.ID,
        Status: "sent",
    }
}
