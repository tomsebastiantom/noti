package createtenant

import "getnoti.com/internal/tenants/domain"

// Input/Output Definitions
type CreateTenantInput struct {
    ID             string
    Name           string
    DefaultChannel domain.NotificationChannel
}

type CreateTenantOutput struct {
    Tenant domain.Tenant
}
