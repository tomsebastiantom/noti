package updateusers

import "getnoti.com/internal/tenants/domain"

type UpdateUsersRequest struct {
    UserID        string
    TenantID      string
    Email         string
    PhoneNumber   string
    DeviceID      string
    WebPushToken  string
    Consents      map[domain.NotificationChannel]bool
    PreferredMode domain.NotificationChannel
}

type UpdateUsersResponse struct {
    Success bool
    Message string
}
