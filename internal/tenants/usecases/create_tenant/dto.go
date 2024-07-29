package createtenant

import "getnoti.com/internal/tenants/domain"

type CreateTenantRequest struct {
    ID             string
    Name           string
    DBConfigs map[string]*domain.DBCredentials
}



type CreateTenantResponse struct {
    Tenant domain.Tenant
}
