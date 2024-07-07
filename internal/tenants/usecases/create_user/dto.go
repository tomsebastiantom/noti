package createuser

import (
   
    "getnoti.com/internal/tenants/domain"
    
)
type CreateUserRequest struct {
    ID           string
    TenantID     string
    Email        string
    PhoneNumber  string
    DeviceID     string
    WebPushToken string
    Consents     map[string]bool
    PreferredMode string
}

type CreateUserResponse struct {
    User domain.User
}
