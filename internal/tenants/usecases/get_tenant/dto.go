package gettenant

import "getnoti.com/internal/tenants/domain"

type GetTenantRequest struct {
    TenantID string
}

type GetTenantResponse struct {
    Tenant domain.Tenant
}
