package createtenant

import "getnoti.com/internal/tenants/domain"


type CreateTenantRequest struct {
    ID             string
    Name           string
    DefaultChannel domain.NotificationChannel
}

type CreateTenantResponse struct {
    Tenant domain.Tenant
}
