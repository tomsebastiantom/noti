package repository

import (
    "context"
    "getnoti.com/internal/tenants/domain"
)

type TenantRepository interface {
    CreateTenant(ctx context.Context, tenant domain.Tenant) error
    GetTenantByID(ctx context.Context, tenantid string) (tenant domain.Tenant, error error)
    Update(ctx context.Context, tenant domain.Tenant) error
    GetAllTenants(ctx context.Context) (tenants []domain.Tenant, error error)
    GetPreferenceByChannel(ctx context.Context, tenantID string, channel string)(map[string]string, error)
}
