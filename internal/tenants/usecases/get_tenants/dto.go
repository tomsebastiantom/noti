package gettenants

import "getnoti.com/internal/tenants/domain"

type GetTenantsRequest struct {
    TenantID string
}

type GetTenantsResponse struct {
    Tenants []domain.Tenant
}
