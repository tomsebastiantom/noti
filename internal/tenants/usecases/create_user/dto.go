package usecase

import (
   
    "getnoti.com/internal/tenants/domain"
    
)
type CreateUserInput struct {
    ID           string
    TenantID     string
    Email        string
    PhoneNumber  string
    DeviceID     string
    WebPushToken string
    Consents     map[domain.NotificationChannel]bool
    PreferredMode domain.NotificationChannel
}

type CreateUserOutput struct {
    User domain.User
}
