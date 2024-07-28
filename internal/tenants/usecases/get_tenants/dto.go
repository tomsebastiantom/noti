package gettenants

import "getnoti.com/internal/tenants/domain"

type GetTenantsRequest struct {
  
}

type GetTenantsResponse struct {
    Tenants []domain.Tenant
}
