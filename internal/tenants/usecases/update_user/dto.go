package updateuser

import "getnoti.com/internal/tenants/domain"

type UpdateUserRequest struct {
    UserID        string
    TenantID      string
    Email         string
    PhoneNumber   string
    DeviceID      string
    WebPushToken  string
    Consents      map[domain.NotificationChannel]bool
    PreferredMode domain.NotificationChannel
}

type UpdateUserResponse struct {
    Success bool
    Message string
}
